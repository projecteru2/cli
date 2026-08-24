package pod

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"google.golang.org/grpc"
)

func TestListPodNodesReturnsStreamError(t *testing.T) {
	want := errors.New("stream broke")
	tests := []struct {
		name   string
		filter string
	}{
		{name: "up", filter: up},
		{name: "all", filter: all},
		{name: "down", filter: down},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &listPodNodesOptions{
				client: &fakePodClient{nodes: func() *fakeStream[corepb.Node] {
					return &fakeStream[corepb.Node]{msgs: []*corepb.Node{{Name: "n1"}}, err: want}
				}},
				filter: tt.filter,
			}

			err := discardStdout(t, func() error { return o.run(t.Context()) })
			if !errors.Is(err, want) {
				t.Errorf("got %v, want %v", err, want)
			}
		})
	}
}

func TestPodResourceReturnsStreamError(t *testing.T) {
	want := errors.New("stream broke")
	o := &resourcePodOptions{
		client: &fakePodClient{resources: func() *fakeStream[corepb.NodeResource] {
			return &fakeStream[corepb.NodeResource]{err: want}
		}},
		name: "dev",
	}

	err := discardStdout(t, func() error { return o.run(t.Context()) })
	if !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestPodResourceRejectsUnparsableFilter(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{name: "empty means no filter", expr: ""},
		{name: "valid", expr: "cpu>40%"},
		{name: "unknown resource", expr: "cpuu>40%", wantErr: true},
		{name: "trailing junk", expr: "cpu>40a", wantErr: true},
		{name: "not an expression", expr: "all", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &resourcePodOptions{expr: tt.expr}
			_, err := o.filter(make(chan *corepb.NodeResource))
			if (err != nil) != tt.wantErr {
				t.Errorf("got %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func discardStdout(t *testing.T, f func() error) error {
	t.Helper()
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer func() { _ = devnull.Close() }()

	orig := os.Stdout
	os.Stdout = devnull
	defer func() { os.Stdout = orig }()
	return f()
}

type fakePodClient struct {
	corepb.CoreRPCClient
	nodes     func() *fakeStream[corepb.Node]
	resources func() *fakeStream[corepb.NodeResource]
}

func (f *fakePodClient) ListPodNodes(context.Context, *corepb.ListNodesOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.Node], error) {
	return f.nodes(), nil
}

func (f *fakePodClient) GetPodResource(context.Context, *corepb.GetPodOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.NodeResource], error) {
	return f.resources(), nil
}

type fakeStream[T any] struct {
	grpc.ClientStream
	msgs []*T
	err  error
	next int
}

func (f *fakeStream[T]) Recv() (*T, error) {
	if f.next >= len(f.msgs) {
		if f.err != nil {
			return nil, f.err
		}
		return nil, io.EOF
	}
	msg := f.msgs[f.next]
	f.next++
	return msg, nil
}
