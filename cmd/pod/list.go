package pod

import (
	"context"
	"fmt"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type listPodsOptions struct {
	client corepb.CoreRPCClient
}

func (o *listPodsOptions) run(ctx context.Context) error {
	resp, err := o.client.ListPods(ctx, &corepb.Empty{})
	if err != nil {
		return fmt.Errorf("[ListPods] send request failed %v", err)
	}

	describe.Pods(resp.GetPods()...)
	return nil
}

func cmdPodList(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &listPodsOptions{
		client: client,
	}
	return o.run(c.Context)
}
