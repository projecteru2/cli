package lambda

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"slices"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"

	"github.com/projecteru2/cli/cmd/utils"
)

func TestLambdaEOFRequiresSynchronousExitCode(t *testing.T) {
	tests := []struct {
		name    string
		async   bool
		wantErr bool
	}{
		{name: "synchronous", wantErr: true},
		{name: "asynchronous", async: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &runLambdaOptions{
				client: &fakeLambdaClient{},
				opts:   &corepb.RunAndWaitOptions{Async: tt.async},
				count:  1,
			}
			_, err := o.lambda(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("got %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateLambdaOptions(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantStorageReq  int64
		wantStorageLim  int64
		wantVolumesReq  []string
		wantVolumesLim  []string
		wantCommands    []string
		wantMemoryLimit int64
	}{
		{
			name:            "defaults",
			args:            []string{"lambda", "echo", "hi"},
			wantCommands:    []string{"echo", "hi"},
			wantMemoryLimit: 512 * 1024 * 1024,
		},
		{
			name: "storage and volumes",
			args: []string{
				"lambda",
				"--storage-request", "1G",
				"--storage", "2G",
				"--volume-request", "AUTO:/data:rw:1G",
				"--volume", "AUTO:/data:rw:2G",
				"echo", "hi",
			},
			wantStorageReq:  1 << 30,
			wantStorageLim:  2 << 30,
			wantVolumesReq:  []string{"AUTO:/data:rw:1G"},
			wantVolumesLim:  []string{"AUTO:/data:rw:2G"},
			wantCommands:    []string{"echo", "hi"},
			wantMemoryLimit: 512 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := runLambdaCommand(t, tt.args)

			storage := decodeParams(t, opts.DeployOptions.Resources[utils.ResourceStorage])
			if got := storage.Int64("storage-request"); got != tt.wantStorageReq {
				t.Errorf("storage-request: got %d, want %d", got, tt.wantStorageReq)
			}
			if got := storage.Int64("storage-limit"); got != tt.wantStorageLim {
				t.Errorf("storage-limit: got %d, want %d", got, tt.wantStorageLim)
			}
			if got := storage.StringSlice("volumes-request"); !slices.Equal(got, tt.wantVolumesReq) {
				t.Errorf("volumes-request: got %v, want %v", got, tt.wantVolumesReq)
			}
			if got := storage.StringSlice("volumes-limit"); !slices.Equal(got, tt.wantVolumesLim) {
				t.Errorf("volumes-limit: got %v, want %v", got, tt.wantVolumesLim)
			}

			cpumem := decodeParams(t, opts.DeployOptions.Resources[utils.ResourceCPUMem])
			if got := cpumem.Int64("memory-limit"); got != tt.wantMemoryLimit {
				t.Errorf("memory-limit: got %d, want %d", got, tt.wantMemoryLimit)
			}
			if got := opts.DeployOptions.Entrypoint.Commands; !slices.Equal(got, tt.wantCommands) {
				t.Errorf("commands: got %v, want %v", got, tt.wantCommands)
			}
		})
	}
}

func TestGenerateLambdaOptionsExtraResources(t *testing.T) {
	opts := runLambdaCommand(t, []string{"lambda", "--extra-resources", `{"gpu":{"count":1}}`, "echo", "hi"})

	gpu, ok := opts.DeployOptions.Resources["gpu"]
	if !ok {
		t.Fatalf("got %v, want a gpu plugin", slices.Sorted(maps.Keys(opts.DeployOptions.Resources)))
	}
	if got := decodeParams(t, gpu).Int64("count"); got != 1 {
		t.Errorf("count: got %d, want 1", got)
	}
}

func TestGenerateLambdaOptionsWithoutCommands(t *testing.T) {
	c := Command()
	c.Action = func(_ context.Context, cmd *cli.Command) error {
		if _, err := generateLambdaOptions(cmd); err == nil {
			t.Error("got nil, want an error for a lambda without commands")
		}
		return nil
	}
	if err := c.Run(t.Context(), []string{"lambda"}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func runLambdaCommand(t *testing.T, args []string) *corepb.RunAndWaitOptions {
	t.Helper()
	var opts *corepb.RunAndWaitOptions
	c := Command()
	c.Action = func(_ context.Context, cmd *cli.Command) error {
		var err error
		opts, err = generateLambdaOptions(cmd)
		return err
	}
	if err := c.Run(t.Context(), args); err != nil {
		t.Fatalf("run: %v", err)
	}
	if opts == nil {
		t.Fatal("got nil options")
	}
	return opts
}

func decodeParams(t *testing.T, raw []byte) resourcetypes.RawParams {
	t.Helper()
	params := resourcetypes.RawParams{}
	if err := json.Unmarshal(raw, &params); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return params
}

type fakeLambdaClient struct {
	corepb.CoreRPCClient
}

func (f *fakeLambdaClient) RunAndWait(context.Context, ...grpc.CallOption) (grpc.BidiStreamingClient[corepb.RunAndWaitOptions, corepb.AttachWorkloadMessage], error) {
	return &fakeLambdaStream{}, nil
}

type fakeLambdaStream struct {
	grpc.ClientStream
}

func (f *fakeLambdaStream) Send(*corepb.RunAndWaitOptions) error {
	return nil
}

func (f *fakeLambdaStream) Recv() (*corepb.AttachWorkloadMessage, error) {
	return nil, io.EOF
}

func (f *fakeLambdaStream) CloseSend() error {
	return nil
}
