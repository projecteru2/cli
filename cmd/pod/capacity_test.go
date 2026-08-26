package pod

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

func TestCapacityResourcesUseRequestAndLimitKeys(t *testing.T) {
	c := Command()
	idx := slices.IndexFunc(c.Commands, func(cmd *cli.Command) bool { return cmd.Name == "capacity" })
	if idx < 0 {
		t.Fatal("capacity command not found")
	}
	var resources map[string][]byte
	c.Commands[idx].Action = func(_ context.Context, cmd *cli.Command) error {
		var err error
		resources, err = capacityResources(cmd)
		return err
	}
	if err := c.Run(t.Context(), []string{"pod", "capacity", "--cpu", "2", "--memory", "4G", "--storage", "1G", "dev"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	cpumem := decodeCapacityParams(t, resources[utils.ResourceCPUMem])
	if cpumem.Float64("cpu-request") != 2 || cpumem.Float64("cpu-limit") != 2 {
		t.Errorf("cpu: got %v, want request and limit 2", cpumem)
	}
	wantRAM := float64(4 * 1024 * 1024 * 1024)
	if cpumem.Float64("memory-request") != wantRAM || cpumem.Float64("memory-limit") != wantRAM {
		t.Errorf("memory: got %v, want request and limit 4G", cpumem)
	}

	storage := decodeCapacityParams(t, resources[utils.ResourceStorage])
	wantStorage := float64(1024 * 1024 * 1024)
	if storage.Float64("storage-request") != wantStorage || storage.Float64("storage-limit") != wantStorage {
		t.Errorf("storage: got %v, want request and limit 1G", storage)
	}
}

func decodeCapacityParams(t *testing.T, raw []byte) resourcetypes.RawParams {
	t.Helper()
	params := resourcetypes.RawParams{}
	if err := json.Unmarshal(raw, &params); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return params
}
