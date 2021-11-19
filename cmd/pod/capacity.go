package pod

import (
	"context"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/google/uuid"
	"github.com/juju/errors"
	"github.com/urfave/cli/v2"
)

type capacityPodOptions struct {
	client       corepb.CoreRPCClient
	podname      string   // podname
	nodenames    []string // node white list
	resourceOpts map[string]*corepb.RawParam
}

func (o *capacityPodOptions) run(ctx context.Context) error {
	opts := &corepb.DeployOptions{
		// resource definitions
		ResourceOpts: o.resourceOpts,
		// deploy options
		Entrypoint: &corepb.EntrypointOptions{
			Name: uuid.New().String(),
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

func cmdPodCapacity(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	stringFlags := []string{"cpu", "memory", "storage"}
	boolFlags := []string{"cpu-bind"}
	resourceOpts := utils.GetResourceOpts(c, stringFlags, nil, boolFlags, nil)

	o := &capacityPodOptions{
		client:       client,
		podname:      name,
		nodenames:    c.StringSlice("nodename"),
		resourceOpts: resourceOpts,
	}
	return o.run(c.Context)
}
