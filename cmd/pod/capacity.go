package pod

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type capacityPodOptions struct {
	client    corepb.CoreRPCClient
	podname   string
	nodenames []string
	resources map[string][]byte
}

func (o *capacityPodOptions) run(ctx context.Context) error {
	opts := &corepb.DeployOptions{
		Resources: o.resources,
		Entrypoint: &corepb.EntrypointOptions{
			Name: strings.ToLower(rand.Text()),
		},
		DeployStrategy: corepb.DeployOptions_DUMMY,
		Podname:        o.podname,
		NodeFilter: &corepb.NodeFilter{
			Includes: o.nodenames,
		},
	}

	resp, err := o.client.CalculateCapacity(ctx, opts)
	if err != nil {
		return err
	}

	describe.PodCapacity(resp.Total, resp.NodeCapacities)
	return nil
}

func cmdPodCapacity(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("pod name must be given")
	}

	memory, err := utils.ParseRAMInHuman(cmd.String(flagMemory))
	if err != nil {
		return fmt.Errorf("parse memory: %w", err)
	}

	storage, err := utils.ParseRAMInHuman(cmd.String(flagStorage))
	if err != nil {
		return fmt.Errorf("parse storage: %w", err)
	}

	cpumem := resourcetypes.RawParams{
		flagCPU:    cmd.Float64(flagCPU),
		flagMemory: memory,
	}
	if cmd.Bool("cpu-bind") {
		cpumem["cpu-bind"] = true
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: resourcetypes.RawParams{flagStorage: storage},
	})
	if err != nil {
		return err
	}

	o := &capacityPodOptions{
		client:    client,
		podname:   name,
		nodenames: cmd.StringSlice("node"),
		resources: resources,
	}
	return o.run(ctx)
}
