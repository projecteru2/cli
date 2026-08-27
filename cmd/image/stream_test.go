package image

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

func TestBuildImageReportsFailure(t *testing.T) {
	tests := []struct {
		name     string
		msg      *corepb.BuildImageMessage
		wantCode int
		wantMsg  string
	}{
		{
			name:     "error without a detail",
			msg:      &corepb.BuildImageMessage{Error: "push image failed"},
			wantCode: -1,
			wantMsg:  "push image failed",
		},
		{
			name: "error with a detail",
			msg: &corepb.BuildImageMessage{
				Error:       "build failed",
				ErrorDetail: &corepb.ErrorDetail{Code: 7, Message: "no space left on device"},
			},
			wantCode: 7,
			wantMsg:  "no space left on device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &buildImageOptions{
				client: &fakeImageClient{build: &fakeStream[corepb.BuildImageMessage]{msgs: []*corepb.BuildImageMessage{tt.msg}}},
				opts:   &corepb.BuildImageOptions{},
			}

			err := o.run(t.Context())
			exitErr, ok := err.(cli.ExitCoder)
			if !ok {
				t.Fatalf("got %v, want a cli.ExitCoder", err)
			}
			if exitErr.ExitCode() != tt.wantCode {
				t.Errorf("code: got %d, want %d", exitErr.ExitCode(), tt.wantCode)
			}
			if exitErr.Error() != tt.wantMsg {
				t.Errorf("message: got %q, want %q", exitErr.Error(), tt.wantMsg)
			}
		})
	}
}

func TestBuildImageEndsEveryProgressLine(t *testing.T) {
	o := &buildImageOptions{
		client: &fakeImageClient{build: &fakeStream[corepb.BuildImageMessage]{msgs: []*corepb.BuildImageMessage{
			{Id: "layer1", Status: "downloading", Progress: "1/2"},
			{Id: "layer1", Status: "downloading", Progress: "2/2"},
			{Id: "layer2", Status: "extracting", Progress: "1/1"},
		}}},
		opts: &corepb.BuildImageOptions{},
	}

	got := captureStdout(t, func() {
		if err := o.run(t.Context()); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	if lines := strings.Count(got, "\n"); lines != 3 {
		t.Errorf("got %d lines in %q, want 3", lines, got)
	}
}

func TestBuildImageKeepsOperatorStatusOutOfStdout(t *testing.T) {
	o := &buildImageOptions{
		client: &fakeImageClient{build: &fakeStream[corepb.BuildImageMessage]{msgs: []*corepb.BuildImageMessage{
			{Stream: "Step 1/2 : FROM alpine\n"},
		}}},
		opts: &corepb.BuildImageOptions{Name: "app"},
	}

	var stdout string
	logged := captureLog(t, func() {
		stdout = captureStdout(t, func() {
			if err := o.run(t.Context()); err != nil {
				t.Fatalf("run: %v", err)
			}
		})
	})

	if strings.Contains(stdout, "complete") {
		t.Errorf("stdout carried operator status: %q", stdout)
	}
	if stdout != "Step 1/2 : FROM alpine\n" {
		t.Errorf("stdout: got %q, want only the remote stream", stdout)
	}
	if !strings.Contains(logged, "build image app complete") {
		t.Errorf("log: got %q, want the completion line", logged)
	}
}

func TestCacheImageReportsFailureReason(t *testing.T) {
	o := &cacheImageOptions{
		client: &fakeImageClient{cache: &fakeStream[corepb.CacheImageMessage]{msgs: []*corepb.CacheImageMessage{
			{Image: "app:v1", Nodename: "node1", Success: false, Message: "no such image"},
		}}},
		images: []string{"app:v1"},
	}

	err := o.run(t.Context())
	if err == nil || !strings.Contains(err.Error(), "no such image") {
		t.Errorf("got %v, want it to carry the failure reason", err)
	}
}

func TestRemoveImageReportsFailure(t *testing.T) {
	o := &removeImageOptions{
		client: &fakeImageClient{remove: &fakeStream[corepb.RemoveImageMessage]{msgs: []*corepb.RemoveImageMessage{
			{Image: "app:v1", Success: false},
		}}},
		images: []string{"app:v1"},
	}

	err := o.run(t.Context())
	if err == nil || !strings.Contains(err.Error(), "remove app:v1 failed") {
		t.Errorf("got %v, want it to report the failed image", err)
	}
}

func TestImageListKeepsPartialResults(t *testing.T) {
	o := &listImageOptions{
		client: &fakeImageClient{list: &fakeStream[corepb.ListImageMessage]{msgs: []*corepb.ListImageMessage{
			{Nodename: "node1", Err: "engine unavailable"},
			{Nodename: "node2"},
		}}},
		opts: &corepb.ListImageOptions{},
	}

	var err error
	_ = captureStdout(t, func() { err = o.run(t.Context()) })
	if err == nil || !strings.Contains(err.Error(), "node1") {
		t.Errorf("got %v, want the failing node named", err)
	}
}

func TestImageCommandsResolveThePodFromTheNode(t *testing.T) {
	tests := []struct {
		name string
		run  func(context.Context, *fakeImageClient) error
	}{
		{
			name: "list",
			run: func(ctx context.Context, c *fakeImageClient) error {
				o := &listImageOptions{client: c, opts: &corepb.ListImageOptions{Nodenames: []string{"node1"}}}
				return o.run(ctx)
			},
		},
		{
			name: "cache",
			run: func(ctx context.Context, c *fakeImageClient) error {
				o := &cacheImageOptions{client: c, images: []string{"app:v1"}, nodenames: []string{"node1"}}
				return o.run(ctx)
			},
		},
		{
			name: "remove",
			run: func(ctx context.Context, c *fakeImageClient) error {
				o := &removeImageOptions{client: c, images: []string{"app:v1"}, nodenames: []string{"node1"}}
				return o.run(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeImageClient{
				podname: "dev",
				build:   &fakeStream[corepb.BuildImageMessage]{},
				cache:   &fakeStream[corepb.CacheImageMessage]{},
				remove:  &fakeStream[corepb.RemoveImageMessage]{},
				list:    &fakeStream[corepb.ListImageMessage]{},
			}
			_ = captureStdout(t, func() {
				if err := tt.run(t.Context(), client); err != nil {
					t.Errorf("run: %v", err)
				}
			})
			if client.askedNodename != "node1" {
				t.Errorf("GetNode nodename: got %q, want %q", client.askedNodename, "node1")
			}
			if client.sentPodname != "dev" {
				t.Errorf("podname sent to core: got %q, want %q", client.sentPodname, "dev")
			}
		})
	}
}

func TestImageCommandsSurfaceTheNodeLookupFailure(t *testing.T) {
	client := &fakeImageClient{nodeErr: errors.New("key: /node/node1: entity count invalid")}
	o := &listImageOptions{client: client, opts: &corepb.ListImageOptions{Nodenames: []string{"node1"}}}

	err := o.run(t.Context())
	if err == nil || !strings.Contains(err.Error(), "node1") {
		t.Errorf("got %v, want core's error naming the node", err)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	out := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		out <- string(b)
	}()

	f()

	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return <-out
}

func captureLog(t *testing.T, f func()) string {
	t.Helper()
	if err := log.SetupLog(t.Context(), &coretypes.ServerLogConfig{Level: "info"}, ""); err != nil {
		t.Fatalf("setup log: %v", err)
	}
	buf := &bytes.Buffer{}
	logger := log.GetGlobalLogger()
	*logger = logger.Output(buf)
	f()
	return buf.String()
}

type fakeImageClient struct {
	corepb.CoreRPCClient
	build   *fakeStream[corepb.BuildImageMessage]
	cache   *fakeStream[corepb.CacheImageMessage]
	remove  *fakeStream[corepb.RemoveImageMessage]
	list    *fakeStream[corepb.ListImageMessage]
	podname string
	nodeErr error

	askedNodename string
	sentPodname   string
}

func (f *fakeImageClient) GetNode(_ context.Context, opts *corepb.GetNodeOptions, _ ...grpc.CallOption) (*corepb.Node, error) {
	f.askedNodename = opts.Nodename
	if f.nodeErr != nil {
		return nil, f.nodeErr
	}
	return &corepb.Node{Name: opts.Nodename, Podname: f.podname}, nil
}

func (f *fakeImageClient) ListImage(_ context.Context, opts *corepb.ListImageOptions, _ ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.ListImageMessage], error) {
	f.sentPodname = opts.Podname
	return f.list, nil
}

func (f *fakeImageClient) BuildImage(context.Context, *corepb.BuildImageOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.BuildImageMessage], error) {
	return f.build, nil
}

func (f *fakeImageClient) CacheImage(_ context.Context, opts *corepb.CacheImageOptions, _ ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.CacheImageMessage], error) {
	f.sentPodname = opts.Podname
	return f.cache, nil
}

func (f *fakeImageClient) RemoveImage(_ context.Context, opts *corepb.RemoveImageOptions, _ ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.RemoveImageMessage], error) {
	f.sentPodname = opts.Podname
	return f.remove, nil
}

type fakeStream[T any] struct {
	grpc.ClientStream
	msgs []*T
	next int
}

func (f *fakeStream[T]) Recv() (*T, error) {
	if f.next >= len(f.msgs) {
		return nil, io.EOF
	}
	msg := f.msgs[f.next]
	f.next++
	return msg, nil
}
