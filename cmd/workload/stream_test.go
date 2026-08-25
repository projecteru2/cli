package workload

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
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

func TestDissociateAddsNodeWorkloads(t *testing.T) {
	tests := []struct {
		name    string
		ids     []string
		nodes   []string
		want    []string
		wantErr string
	}{
		{
			name:  "node workloads join the given ids",
			ids:   []string{"cid1"},
			nodes: []string{"node1"},
			want:  []string{"cid1", "cid2", "cid3"},
		},
		{
			name: "ids alone are kept",
			ids:  []string{"cid1"},
			want: []string{"cid1"},
		},
		{
			name:    "an empty node dissociates nothing",
			nodes:   []string{"node2"},
			wantErr: "no workloads found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeWorkloadClient{
				nodeWorkloads: map[string][]string{"node1": {"cid2", "cid3"}},
				dissociate:    &fakeStream[corepb.DissociateWorkloadMessage]{},
			}
			o := &dissociateWorkloadsOptions{client: client, ids: tt.ids, nodes: tt.nodes}

			err := o.run(t.Context())
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got %v, want it to mention %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("got %v, want nil", err)
			}
			if !slices.Equal(client.dissociated, tt.want) {
				t.Errorf("got %v, want %v", client.dissociated, tt.want)
			}
		})
	}
}

func TestDissociateReadsTheNodeFlag(t *testing.T) {
	var nodes []string
	c := Command()
	lookupSubcommand(t, c, "dissociate").Action = func(_ context.Context, cmd *cli.Command) error {
		nodes = cmd.StringSlice(flagNode)
		return nil
	}
	if err := c.Run(t.Context(), []string{"workload", "dissociate", "--node", "node1", "--node", "node2"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !slices.Equal(nodes, []string{"node1", "node2"}) {
		t.Errorf("got %v, want node1 and node2", nodes)
	}
}

func TestFailedItemsReachTheExitCode(t *testing.T) {
	tests := []struct {
		name    string
		run     func(context.Context) error
		wantErr string
	}{
		{
			name: "create",
			run: func(ctx context.Context) error {
				client := &fakeWorkloadClient{create: &fakeStream[corepb.CreateWorkloadMessage]{msgs: []*corepb.CreateWorkloadMessage{
					{Nodename: "node1", Success: false, Error: "not enough resource"},
				}}}
				return doCreateWorkload(ctx, client, &corepb.DeployOptions{})
			},
			wantErr: "not enough resource",
		},
		{
			name: "replace",
			run: func(ctx context.Context) error {
				client := &fakeWorkloadClient{replace: &fakeStream[corepb.ReplaceWorkloadMessage]{msgs: []*corepb.ReplaceWorkloadMessage{
					{Error: "image not found", Remove: &corepb.RemoveWorkloadMessage{Id: "cid1"}},
				}}}
				return doReplaceWorkload(ctx, client, &corepb.DeployOptions{}, false, nil, nil)
			},
			wantErr: "image not found",
		},
		{
			name: "send",
			run: func(ctx context.Context) error {
				o := &sendWorkloadsOptions{client: &fakeWorkloadClient{sendFiles: &fakeStream[corepb.SendMessage]{msgs: []*corepb.SendMessage{
					{Id: "cid1", Path: "/etc/app", Error: "read only file system"},
				}}}}
				return o.run(ctx)
			},
			wantErr: "read only file system",
		},
		{
			name: "control",
			run: func(ctx context.Context) error {
				o := &controlWorkloadsOptions{
					client: &fakeWorkloadClient{control: &fakeStream[corepb.ControlWorkloadMessage]{msgs: []*corepb.ControlWorkloadMessage{
						{Id: "cid1", Error: "hook refused"},
					}}},
					action: "stop",
				}
				return o.run(ctx)
			},
			wantErr: "hook refused",
		},
		{
			name: "remove",
			run: func(ctx context.Context) error {
				o := &removeWorkloadsOptions{client: &fakeWorkloadClient{remove: &fakeStream[corepb.RemoveWorkloadMessage]{msgs: []*corepb.RemoveWorkloadMessage{
					{Id: "cid1", Success: false},
				}}}}
				return o.run(ctx)
			},
			wantErr: "remove cid1 failed",
		},
		{
			name: "dissociate",
			run: func(ctx context.Context) error {
				o := &dissociateWorkloadsOptions{
					client: &fakeWorkloadClient{dissociate: &fakeStream[corepb.DissociateWorkloadMessage]{msgs: []*corepb.DissociateWorkloadMessage{
						{Id: "cid1", Error: "still running"},
					}}},
					ids: []string{"cid1"},
				}
				return o.run(ctx)
			},
			wantErr: "still running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run(t.Context())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestSendLargeRejectsAnEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := runCommand(t, "sendlarge", "--file", path+":/etc/app", "cid1")
	if err == nil || !strings.Contains(err.Error(), "nothing to send") {
		t.Errorf("got %v, want it to refuse the empty file", err)
	}
}

func TestCopyRejectsAMalformedSource(t *testing.T) {
	err := runCommand(t, "copy", "cid1")
	if err == nil || !strings.Contains(err.Error(), "invalid source") {
		t.Errorf("got %v, want it to name the malformed source", err)
	}
}

func runCommand(t *testing.T, args ...string) error {
	t.Helper()
	exiter, writer := cli.OsExiter, cli.ErrWriter
	cli.OsExiter = func(int) {}
	cli.ErrWriter = io.Discard
	t.Cleanup(func() { cli.OsExiter, cli.ErrWriter = exiter, writer })
	return Command().Run(t.Context(), append([]string{"workload"}, args...))
}

type fakeWorkloadClient struct {
	corepb.CoreRPCClient
	copy          *fakeStream[corepb.CopyMessage]
	send          *fakeSendStream
	sendFiles     *fakeStream[corepb.SendMessage]
	create        *fakeStream[corepb.CreateWorkloadMessage]
	replace       *fakeStream[corepb.ReplaceWorkloadMessage]
	control       *fakeStream[corepb.ControlWorkloadMessage]
	remove        *fakeStream[corepb.RemoveWorkloadMessage]
	dissociate    *fakeStream[corepb.DissociateWorkloadMessage]
	nodeWorkloads map[string][]string
	dissociated   []string
}

func (f *fakeWorkloadClient) CreateWorkload(context.Context, *corepb.DeployOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.CreateWorkloadMessage], error) {
	return f.create, nil
}

func (f *fakeWorkloadClient) ReplaceWorkload(context.Context, *corepb.ReplaceOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.ReplaceWorkloadMessage], error) {
	return f.replace, nil
}

func (f *fakeWorkloadClient) Send(context.Context, *corepb.SendOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.SendMessage], error) {
	return f.sendFiles, nil
}

func (f *fakeWorkloadClient) ControlWorkload(context.Context, *corepb.ControlWorkloadOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.ControlWorkloadMessage], error) {
	return f.control, nil
}

func (f *fakeWorkloadClient) RemoveWorkload(context.Context, *corepb.RemoveWorkloadOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.RemoveWorkloadMessage], error) {
	return f.remove, nil
}

func (f *fakeWorkloadClient) Copy(context.Context, *corepb.CopyOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.CopyMessage], error) {
	return f.copy, nil
}

func (f *fakeWorkloadClient) DissociateWorkload(_ context.Context, opts *corepb.DissociateWorkloadOptions, _ ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.DissociateWorkloadMessage], error) {
	f.dissociated = opts.IDs
	return f.dissociate, nil
}

func (f *fakeWorkloadClient) ListNodeWorkloads(_ context.Context, opts *corepb.GetNodeOptions, _ ...grpc.CallOption) (*corepb.Workloads, error) {
	workloads := &corepb.Workloads{}
	for _, id := range f.nodeWorkloads[opts.Nodename] {
		workloads.Workloads = append(workloads.Workloads, &corepb.Workload{Id: id})
	}
	return workloads, nil
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
