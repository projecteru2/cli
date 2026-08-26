package node

import (
	"context"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

func TestGenerateSetNodeOptions(t *testing.T) {
	var opts *corepb.SetNodeOptions
	c := Command()
	lookupSubcommand(t, c, "set").Action = func(_ context.Context, cmd *cli.Command) error {
		var err error
		opts, err = generateSetNodeOptions(cmd)
		return err
	}
	if err := c.Run(t.Context(), []string{"node", "set", "--cpu", "+2", "--rm-disk", "sda", "node1"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if opts == nil {
		t.Fatal("got nil options")
	}

	cpumem := decodeParams(t, opts.Resources[utils.ResourceCPUMem])
	if got := cpumem.String("cpu"); got != "+2" {
		t.Errorf("cpu: got %q, want %q", got, "+2")
	}
	storage := decodeParams(t, opts.Resources[utils.ResourceStorage])
	if got := storage.String("rm-disks"); got != "sda" {
		t.Errorf("rm-disks: got %q, want %q", got, "sda")
	}
}
