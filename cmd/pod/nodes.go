package pod

import (
	"context"
	"errors"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type listPodNodesOptions struct {
	client          corepb.CoreRPCClient
	name            string
	filter          string
	labels          map[string]string
	timeoutInSecond int32
	showInfo        bool
	stream          bool
}

func (o *listPodNodesOptions) run(ctx context.Context) error {
	ch, wait, err := o.listChan(ctx, &corepb.ListNodesOptions{
		Podname:         o.name,
		All:             o.filter != up,
		Labels:          o.labels,
		TimeoutInSecond: o.timeoutInSecond,
		SkipInfo:        !o.showInfo,
	})
	if err != nil {
		return err
	}
	if o.filter == down {
		ch = downOnly(ch)
	}
	describe.NodesStream(ch, o.showInfo, o.stream)
	return wait()
}

func (o *listPodNodesOptions) listChan(ctx context.Context, opt *corepb.ListNodesOptions) (<-chan *corepb.Node, func() error, error) {
	stream, err := o.client.ListPodNodes(ctx, opt)
	if err != nil {
		return nil, nil, err
	}
	ch, wait := utils.StreamToChan(stream.Recv)
	return ch, wait, nil
}

func cmdPodListNodes(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	filter := strings.ToLower(cmd.String("filter"))
	if filter != up && filter != down && filter != all {
		return errors.New("filter should be one of up/down/all")
	}

	o := &listPodNodesOptions{
		client:          client,
		name:            cmd.Args().First(),
		filter:          filter,
		labels:          utils.SplitEquality(cmd.StringSlice("label")),
		timeoutInSecond: int32(cmd.Int("timeout")), //nolint:gosec
		showInfo:        cmd.Bool("show-info"),
		stream:          cmd.Bool("stream"),
	}
	return o.run(ctx)
}

func downOnly(nodes <-chan *corepb.Node) <-chan *corepb.Node {
	down := make(chan *corepb.Node)
	go func() {
		defer close(down)
		for node := range nodes {
			if !node.Available || node.Bypass {
				down <- node
			}
		}
	}()
	return down
}
