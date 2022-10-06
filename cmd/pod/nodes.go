package pod

import (
	"context"
	"io"
	"strings"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/juju/errors"
	"github.com/urfave/cli/v2"
)

const (
	up   = "up"
	down = "down"
	all  = "all"
)

type listPodNodesOptions struct {
	client          corepb.CoreRPCClient
	name            string
	filter          string
	labels          map[string]string
	timeoutInSecond int32
	showInfo        bool
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

	o.describeNodes(describe.ToNodeChan(unavailNodes...), true)
	return nil
}

func (o *listPodNodesOptions) listUpOrAll(ctx context.Context) error {
	// filter == all, list all nodes
	// filter == up, list available nodes only
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

	o.describeNodes(ch, true)

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

		for {
			node, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					println(err.Error())
				}
				return
			}

			ch <- node
		}
	}()

	return ch, nil
}

func (o *listPodNodesOptions) describeNodes(nodes <-chan *corepb.Node, stream bool) {
	if o.showInfo {
		describe.NodesWithInfo(nodes, stream)
	} else {
		describe.Nodes(nodes, stream)
	}
}

func cmdPodListNodes(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	filter := strings.ToLower(c.String("filter"))
	if !(filter == up || filter == down || filter == all) {
		return errors.New("filter should be one of up/down/all")
	}

	o := &listPodNodesOptions{
		client:          client,
		name:            c.Args().First(),
		filter:          filter,
		labels:          utils.SplitEquality(c.StringSlice("label")),
		timeoutInSecond: int32(c.Int("timeout")),
		showInfo:        c.Bool("show-info"),
	}
	return o.run(c.Context)
}
