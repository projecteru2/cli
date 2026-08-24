package pod

import (
	"context"
	"fmt"

	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"

	"github.com/urfave/cli/v3"
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

func cmdPodList(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &listPodsOptions{
		client: client,
	}
	return o.run(ctx)
}
