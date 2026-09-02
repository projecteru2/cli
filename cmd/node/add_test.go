package node

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

func TestGenerateAddNodeOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCPU     int64
		wantShare   int64
		wantMemory  string
		wantStorage string
		wantVolumes []string
	}{
		{
			name:      "cpu and share",
			args:      []string{"node", "add", "--endpoint", "process://127.0.0.1", "--cpu", "64", "--share", "100", "dev"},
			wantCPU:   64,
			wantShare: 100,
		},
		{
			name:        "memory storage and volumes",
			args:        []string{"node", "add", "--endpoint", "process://127.0.0.1", "--memory", "1G", "--storage", "10G", "--volume", "/data:100G", "dev"},
			wantMemory:  "1G",
			wantStorage: "10G",
			wantVolumes: []string{"/data:100G"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := runAddNodeCommand(t, tt.args)
			if opts.Podname != "dev" {
				t.Errorf("podname: got %q, want %q", opts.Podname, "dev")
			}

			cpumem := decodeParams(t, opts.Resources[utils.ResourceCPUMem])
			if got := cpumem.Int64("cpu"); got != tt.wantCPU {
				t.Errorf("cpu: got %d, want %d", got, tt.wantCPU)
			}
			if got := cpumem.Int64("share"); got != tt.wantShare {
				t.Errorf("share: got %d, want %d", got, tt.wantShare)
			}
			if got := cpumem.String("memory"); got != tt.wantMemory {
				t.Errorf("memory: got %q, want %q", got, tt.wantMemory)
			}

			storage := decodeParams(t, opts.Resources[utils.ResourceStorage])
			if got := storage.String(flagStorage); got != tt.wantStorage {
				t.Errorf("storage: got %q, want %q", got, tt.wantStorage)
			}
			if got := storage.StringSlice("volumes"); !slices.Equal(got, tt.wantVolumes) {
				t.Errorf("volumes: got %v, want %v", got, tt.wantVolumes)
			}
		})
	}
}

func TestGenerateAddNodeOptionsRequiresEndpoint(t *testing.T) {
	c := Command()
	lookupSubcommand(t, c, "add").Action = func(_ context.Context, cmd *cli.Command) error {
		if _, err := generateAddNodeOptions(cmd); err == nil {
			t.Error("got nil, want an error for a missing endpoint")
		}
		return nil
	}
	if err := c.Run(t.Context(), []string{"node", "add", "dev"}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestGenerateAddNodeOptionsWithoutPod(t *testing.T) {
	c := Command()
	lookupSubcommand(t, c, "add").Action = func(_ context.Context, cmd *cli.Command) error {
		if _, err := generateAddNodeOptions(cmd); err == nil {
			t.Error("got nil, want an error for node add without a pod")
		}
		return nil
	}
	if err := c.Run(t.Context(), []string{"node", "add"}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func runAddNodeCommand(t *testing.T, args []string) *corepb.AddNodeOptions {
	t.Helper()
	var opts *corepb.AddNodeOptions
	c := Command()
	add := lookupSubcommand(t, c, "add")
	add.Action = func(_ context.Context, cmd *cli.Command) error {
		var err error
		opts, err = generateAddNodeOptions(cmd)
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

func lookupSubcommand(t *testing.T, cmd *cli.Command, name string) *cli.Command {
	t.Helper()
	idx := slices.IndexFunc(cmd.Commands, func(c *cli.Command) bool { return c.Name == name })
	if idx < 0 {
		t.Fatalf("subcommand %q not found", name)
	}
	return cmd.Commands[idx]
}

func decodeParams(t *testing.T, raw []byte) resourcetypes.RawParams {
	t.Helper()
	params := resourcetypes.RawParams{}
	if len(raw) == 0 {
		return params
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return params
}
