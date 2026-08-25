package workload

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"google.golang.org/grpc"
)

var errStreamBroke = errors.New("stream broke")

func TestCopyReturnsFailures(t *testing.T) {
	tests := []struct {
		name    string
		msgs    []*corepb.CopyMessage
		wantErr string
	}{
		{
			name:    "core reports a failure",
			msgs:    []*corepb.CopyMessage{{Id: "cid1", Name: "app", Error: "no such path"}},
			wantErr: "no such path",
		},
		{
			name: "everything is written",
			msgs: []*corepb.CopyMessage{{Id: "cid1", Name: "app", Data: []byte("payload")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			o := &copyWorkloadsOptions{
				client:          &fakeWorkloadClient{copy: &fakeStream[corepb.CopyMessage]{msgs: tt.msgs}},
				dir:             dir,
				pathsByWorkload: map[string][]string{"cid1": {"/etc/app"}},
			}

			err := o.run(t.Context())
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("got %v, want nil", err)
				}
				entries, readErr := os.ReadDir(dir)
				if readErr != nil {
					t.Fatalf("read dir: %v", readErr)
				}
				if len(entries) != 1 {
					t.Errorf("got %d files, want 1", len(entries))
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestSendLargeReturnsFailures(t *testing.T) {
	tests := []struct {
		name    string
		msgs    []*corepb.SendMessage
		recvErr error
		wantErr string
	}{
		{
			name:    "core reports a failure",
			msgs:    []*corepb.SendMessage{{Id: "cid1", Path: "/etc/app", Error: "read only file system"}},
			wantErr: "read only file system",
		},
		{
			name:    "the stream breaks",
			recvErr: errStreamBroke,
			wantErr: errStreamBroke.Error(),
		},
		{
			name: "every chunk lands",
			msgs: []*corepb.SendMessage{{Id: "cid1", Path: "/etc/app"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &sendLargeWorkloadsOptions{
				client: &fakeWorkloadClient{send: &fakeSendStream{
					msgs: tt.msgs,
					err:  tt.recvErr,
				}},
				ids:     []string{"cid1"},
				dst:     "/etc/app",
				content: []byte("payload"),
			}

			err := o.run(t.Context())
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("got %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

type fakeWorkloadClient struct {
	corepb.CoreRPCClient
	copy *fakeStream[corepb.CopyMessage]
	send *fakeSendStream
}

func (f *fakeWorkloadClient) Copy(context.Context, *corepb.CopyOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.CopyMessage], error) {
	return f.copy, nil
}

func (f *fakeWorkloadClient) SendLargeFile(context.Context, ...grpc.CallOption) (grpc.BidiStreamingClient[corepb.FileOptions, corepb.SendMessage], error) {
	return f.send, nil
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

type fakeSendStream struct {
	grpc.ClientStream
	msgs []*corepb.SendMessage
	err  error
	next int
}

func (f *fakeSendStream) Send(*corepb.FileOptions) error {
	return nil
}

func (f *fakeSendStream) Recv() (*corepb.SendMessage, error) {
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

func (f *fakeSendStream) CloseSend() error {
	return nil
}
