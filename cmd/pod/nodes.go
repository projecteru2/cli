package pod

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/projecteru2/core/log"
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
	if o.filter == up || o.filter == all {
		return o.listUpOrAll(ctx)
	}
	return o.listDown(ctx)
}

func (o *listPodNodesOptions) listDown(ctx context.Context) error {
	allNodes, err := o.list(ctx, &corepb.ListNodesOptions{
		Podname:         o.name,
		All:             true,
		Labels:          o.labels,
		TimeoutInSecond: o.timeoutInSecond,
		SkipInfo:        !o.showInfo,
	})
	if err != nil {
		return err
	}

	availNodes, err := o.list(ctx, &corepb.ListNodesOptions{
		Podname:         o.name,
		All:             false,
		Labels:          o.labels,
		TimeoutInSecond: o.timeoutInSecond,
		SkipInfo:        !o.showInfo,
	})
	if err != nil {
		return err
	}

	availableNodes := map[string]*corepb.Node{}
	for _, node := range availNodes {
		availableNodes[node.Name] = node
	}

	unavailNodes := []*corepb.Node{}
	for _, node := range allNodes {
		if _, ok := availableNodes[node.Name]; ok {
			continue
		}
		unavailNodes = append(unavailNodes, node)
	}

	o.describeNodes(describe.ToChan(unavailNodes...))
	return nil
}

func (o *listPodNodesOptions) listUpOrAll(ctx context.Context) error {
	ch, err := o.listChan(ctx, &corepb.ListNodesOptions{
		Podname:         o.name,
		All:             o.filter == all,
		Labels:          o.labels,
		TimeoutInSecond: o.timeoutInSecond,
		SkipInfo:        !o.showInfo,
	})
	if err != nil {
		return err
	}

	o.describeNodes(ch)

	return nil
}

func (o *listPodNodesOptions) list(ctx context.Context, opt *corepb.ListNodesOptions) ([]*corepb.Node, error) {
	ch, err := o.listChan(ctx, opt)
	if err != nil {
		return nil, err
	}

	nodes := []*corepb.Node{}
	for n := range ch {
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (o *listPodNodesOptions) listChan(ctx context.Context, opt *corepb.ListNodesOptions) (<-chan *corepb.Node, error) {
	stream, err := o.client.ListPodNodes(ctx, opt)
	if err != nil {
		return nil, err
	}

	ch := make(chan *corepb.Node)
	go func() {
		defer close(ch)

		logger := log.WithFunc("pod.listPodNodesOptions.listChan")
		for {
			node, recvErr := stream.Recv()
			if recvErr != nil {
				if !errors.Is(recvErr, io.EOF) {
					logger.Error(ctx, recvErr)
				}
				return
			}

			ch <- node
		}
	}()

	return ch, nil
}

func (o *listPodNodesOptions) describeNodes(nodes <-chan *corepb.Node) {
	if o.showInfo {
		describe.NodesWithInfo(nodes, o.stream)
	} else {
		describe.Nodes(nodes, o.stream)
	}
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
