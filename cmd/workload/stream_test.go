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
	"testing/iotest"
	"time"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

var errStreamBroke = errors.New("stream broke")

func TestCopyReturnsFailures(t *testing.T) {
	tests := []struct {
		name    string
		msgs    []*corepb.CopyMessage
		recvErr error
		wantErr string
	}{
		{
			name:    "core reports a failure",
			msgs:    []*corepb.CopyMessage{{Id: "cid1", Path: "/etc/app", Error: "no such path"}},
			wantErr: "no such path",
		},
		{
			name:    "a broken stream keeps the diagnostics",
			msgs:    []*corepb.CopyMessage{{Id: "cid1", Path: "/etc/app", Error: "no such path"}},
			recvErr: errStreamBroke,
			wantErr: "no such path",
		},
		{
			name: "everything is written",
			msgs: []*corepb.CopyMessage{{Id: "cid1", Path: "/etc/app", Data: []byte("payload")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			o := &copyWorkloadsOptions{
				client:          &fakeWorkloadClient{copy: &fakeStream[corepb.CopyMessage]{msgs: tt.msgs, err: tt.recvErr}},
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

func TestCopyKeepsPathsApart(t *testing.T) {
	dir := t.TempDir()
	o := &copyWorkloadsOptions{
		client: &fakeWorkloadClient{copy: &fakeStream[corepb.CopyMessage]{msgs: []*corepb.CopyMessage{
			{Id: "cid1", Path: "/etc/hosts", Data: []byte("hosts")},
			{Id: "cid1", Path: "/a/b", Data: []byte("slash")},
			{Id: "cid1", Path: "/a_b", Data: []byte("underscore")},
		}}},
		dir:             dir,
		pathsByWorkload: map[string][]string{"cid1": {"/etc/hosts", "/a/b", "/a_b"}},
	}

	if err := o.run(t.Context()); err != nil {
		t.Fatalf("run: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("got %d files, want one per path even when names collide", len(entries))
	}
}

func TestCopyKeepsWorkloadsApart(t *testing.T) {
	const (
		firstID  = "1111111111111111111111111abcdef0"
		secondID = "2222222222222222222222222abcdef0"
	)

	dir := t.TempDir()
	o := &copyWorkloadsOptions{
		client: &fakeWorkloadClient{copy: &fakeStream[corepb.CopyMessage]{msgs: []*corepb.CopyMessage{
			{Id: firstID, Path: "/etc/hosts", Data: []byte("first")},
			{Id: secondID, Path: "/etc/hosts", Data: []byte("second")},
		}}},
		dir: dir,
		pathsByWorkload: map[string][]string{
			firstID:  {"/etc/hosts"},
			secondID: {"/etc/hosts"},
		},
	}

	if err := o.run(t.Context()); err != nil {
		t.Fatalf("run: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("got %d files, want one per workload", len(entries))
	}
}

func TestCopyBoundsTheFilename(t *testing.T) {
	deep := "/opt/" + strings.Repeat("depth/", 40)
	id := strings.Repeat("f", 64)
	dir := t.TempDir()
	o := &copyWorkloadsOptions{
		client: &fakeWorkloadClient{copy: &fakeStream[corepb.CopyMessage]{msgs: []*corepb.CopyMessage{
			{Id: id, Path: deep + "index.js", Data: []byte("first")},
			{Id: id, Path: deep + "other.js", Data: []byte("second")},
		}}},
		dir:             dir,
		pathsByWorkload: map[string][]string{id: {deep + "index.js", deep + "other.js"}},
	}

	if err := o.run(t.Context()); err != nil {
		t.Fatalf("run: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d files, want distinct files for paths sharing a truncated prefix", len(entries))
	}
	for _, entry := range entries {
		if len(entry.Name()) > 255 {
			t.Errorf("filename length %d exceeds NAME_MAX", len(entry.Name()))
		}
	}
}

func TestCopyMergesRepeatedIDs(t *testing.T) {
	sources, err := parseCopySources([]string{"cid1:/etc/hosts", "cid1:/etc/passwd,/etc/hosts"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !slices.Equal(sources["cid1"], []string{"/etc/hosts", "/etc/passwd"}) {
		t.Errorf("got %v, want both paths once", sources["cid1"])
	}
}

func TestCopyPreservesPathSemantics(t *testing.T) {
	sources, err := parseCopySources([]string{"cid1:etc/hosts,/etc/hosts"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !slices.Equal(sources["cid1"], []string{"etc/hosts", "/etc/hosts"}) {
		t.Errorf("got %v, want relative and absolute paths unchanged", sources["cid1"])
	}
}

func TestCopyRejectsAnEmptyPath(t *testing.T) {
	for _, source := range []string{"cid1:,/etc/hosts", "cid1:/etc/hosts,", "cid1:/etc/hosts,,/etc/passwd"} {
		if _, err := parseCopySources([]string{source}); err == nil {
			t.Errorf("got nil for %q, want an invalid source error", source)
		}
	}
}

func TestExecSurfacesTheStreamStatusOnSendFailure(t *testing.T) {
	o := &execWorkloadOptions{
		client:   &fakeWorkloadClient{exec: &fakeExecStream{sendErr: io.EOF, recvErr: errStreamBroke}},
		id:       "cid1",
		commands: []string{"date"},
	}

	err := o.run(t.Context())
	if err == nil || !strings.Contains(err.Error(), errStreamBroke.Error()) {
		t.Errorf("got %v, want the stream status", err)
	}
}

func TestSendLargeReturnsFailures(t *testing.T) {
	tests := []struct {
		name    string
		msgs    []*corepb.SendMessage
		recvErr error
		sendErr error
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
			name:    "a send abort surfaces the stream status",
			msgs:    []*corepb.SendMessage{{Id: "cid1", Path: "/etc/app", Error: "workload not found"}},
			sendErr: io.EOF,
			wantErr: "workload not found",
		},
		{
			name:    "a send abort without a status still fails",
			sendErr: io.EOF,
			wantErr: "closed before every chunk landed",
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
					msgs:    tt.msgs,
					err:     tt.recvErr,
					sendErr: tt.sendErr,
				}},
				ids:  []string{"cid1"},
				dst:  "/etc/app",
				src:  strings.NewReader("payload"),
				size: 7,
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

func TestDissociateDeduplicatesSelections(t *testing.T) {
	client := &fakeWorkloadClient{
		nodeWorkloads: map[string][]string{"node1": {"cid1", "cid2", "cid2"}},
		dissociate:    &fakeStream[corepb.DissociateWorkloadMessage]{},
	}
	o := &dissociateWorkloadsOptions{client: client, ids: []string{"cid1", "cid1"}, nodes: []string{"node1"}}

	if err := o.run(t.Context()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !slices.Equal(client.dissociated, []string{"cid1", "cid2"}) {
		t.Errorf("got %v, want cid1 and cid2", client.dissociated)
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

func TestAutoReplaceChecksTheTargetPod(t *testing.T) {
	tests := []struct {
		name        string
		workloads   []*corepb.Workload
		wantCreate  bool
		wantReplace bool
	}{
		{name: "other pod creates", workloads: []*corepb.Workload{{Podname: "prod"}}, wantCreate: true},
		{name: "target pod replaces", workloads: []*corepb.Workload{{Podname: "prod"}, {Podname: "dev"}}, wantReplace: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeWorkloadClient{
				list:    &fakeStream[corepb.Workload]{msgs: tt.workloads},
				create:  &fakeStream[corepb.CreateWorkloadMessage]{},
				replace: &fakeStream[corepb.ReplaceWorkloadMessage]{},
			}
			o := &deployWorkloadsOptions{
				client:      client,
				opts:        &corepb.DeployOptions{Name: "app", Podname: "dev", Entrypoint: &corepb.EntrypointOptions{Name: "web"}},
				autoReplace: true,
			}

			if err := o.run(t.Context()); err != nil {
				t.Fatalf("run: %v", err)
			}
			if client.created != tt.wantCreate || client.replaced != tt.wantReplace {
				t.Errorf("got create=%v replace=%v, want create=%v replace=%v", client.created, client.replaced, tt.wantCreate, tt.wantReplace)
			}
			if client.listOpts.Limit != 0 {
				t.Errorf("got limit %d, want the complete app and entrypoint selection", client.listOpts.Limit)
			}
		})
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
			name: "logs",
			run: func(ctx context.Context) error {
				o := &workloadLogsOptions{client: &fakeWorkloadClient{logs: &fakeStream[corepb.LogStreamMessage]{msgs: []*corepb.LogStreamMessage{
					{Id: "cid1", Error: "engine unavailable"},
				}}}, id: "cid1"}
				return o.run(ctx)
			},
			wantErr: "engine unavailable",
		},
		{
			name: "realloc",
			run: func(ctx context.Context) error {
				o := &reallocWorkloadsOptions{
					client: &fakeWorkloadClient{realloc: &corepb.ReallocResourceMessage{Error: "insufficient cpu"}},
					opts:   &corepb.ReallocOptions{Id: "cid1"},
				}
				return o.run(ctx)
			},
			wantErr: "insufficient cpu",
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

func TestSendLargeCancelsTheStreamOnALocalReadError(t *testing.T) {
	o := &sendLargeWorkloadsOptions{
		client: &hangingSendClient{},
		ids:    []string{"cid1"},
		dst:    "/etc/app",
		src:    iotest.ErrReader(errors.New("is a directory")),
		size:   4096,
	}

	done := make(chan error, 1)
	go func() { done <- o.run(t.Context()) }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "is a directory") {
			t.Errorf("got %v, want the read error", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run waited on the receiver after a local read error")
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
	logs          *fakeStream[corepb.LogStreamMessage]
	list          *fakeStream[corepb.Workload]
	exec          *fakeExecStream
	realloc       *corepb.ReallocResourceMessage
	nodeWorkloads map[string][]string
	dissociated   []string
	listOpts      *corepb.ListWorkloadsOptions
	created       bool
	replaced      bool
}

func (f *fakeWorkloadClient) CreateWorkload(context.Context, *corepb.DeployOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.CreateWorkloadMessage], error) {
	f.created = true
	return f.create, nil
}

func (f *fakeWorkloadClient) ReplaceWorkload(context.Context, *corepb.ReplaceOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.ReplaceWorkloadMessage], error) {
	f.replaced = true
	return f.replace, nil
}

func (f *fakeWorkloadClient) ListWorkloads(_ context.Context, opts *corepb.ListWorkloadsOptions, _ ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.Workload], error) {
	f.listOpts = opts
	return f.list, nil
}

func (f *fakeWorkloadClient) LogStream(context.Context, *corepb.LogStreamOptions, ...grpc.CallOption) (grpc.ServerStreamingClient[corepb.LogStreamMessage], error) {
	return f.logs, nil
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

func (f *fakeWorkloadClient) ExecuteWorkload(context.Context, ...grpc.CallOption) (grpc.BidiStreamingClient[corepb.ExecuteWorkloadOptions, corepb.AttachWorkloadMessage], error) {
	return f.exec, nil
}

func (f *fakeWorkloadClient) ReallocResource(context.Context, *corepb.ReallocOptions, ...grpc.CallOption) (*corepb.ReallocResourceMessage, error) {
	return f.realloc, nil
}

type fakeExecStream struct {
	grpc.ClientStream
	sendErr error
	recvErr error
}

func (f *fakeExecStream) Send(*corepb.ExecuteWorkloadOptions) error {
	return f.sendErr
}

func (f *fakeExecStream) Recv() (*corepb.AttachWorkloadMessage, error) {
	if f.recvErr != nil {
		return nil, f.recvErr
	}
	return nil, io.EOF
}

func (f *fakeExecStream) CloseSend() error {
	return nil
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

type fakeSendStream struct {
	grpc.ClientStream
	msgs    []*corepb.SendMessage
	err     error
	sendErr error
	next    int
}

func (f *fakeSendStream) Send(*corepb.FileOptions) error {
	return f.sendErr
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

type hangingSendClient struct {
	corepb.CoreRPCClient
}

func (h *hangingSendClient) SendLargeFile(ctx context.Context, _ ...grpc.CallOption) (grpc.BidiStreamingClient[corepb.FileOptions, corepb.SendMessage], error) {
	return &hangingSendStream{ctx: ctx}, nil
}

type hangingSendStream struct {
	grpc.ClientStream
	ctx context.Context
}

func (h *hangingSendStream) Send(*corepb.FileOptions) error {
	return nil
}

func (h *hangingSendStream) Recv() (*corepb.SendMessage, error) {
	<-h.ctx.Done()
	return nil, h.ctx.Err()
}

func (h *hangingSendStream) CloseSend() error {
	return nil
}
