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
	stream          bool
}

func (o *listPodNodesOptions) run(ctx context.Context) error {
	if o.filter == up || o.filter == all {
		return o.listUpOrAll(ctx)
	}
	return o.listDown(ctx)
}

func (o *listPodNodesOptions) listDown(ctx context.Context) error {
	resp1, err := o.client.ListPodNodes(ctx, &corepb.ListNodesOptions{
		Podname:         o.name,
		All:             true,
		Labels:          o.labels,
		TimeoutInSecond: o.timeoutInSecond,
		SkipInfo:        !o.showInfo,
	})
	if err != nil {
		return err
	}

	resp2, err := o.client.ListPodNodes(ctx, &corepb.ListNodesOptions{
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
	for _, node := range resp2.GetNodes() {
		availableNodes[node.Name] = node
	}

	nodes := []*corepb.Node{}
	for _, node := range resp1.GetNodes() {
		if _, ok := availableNodes[node.Name]; ok {
			continue
		}
		nodes = append(nodes, node)
	}

	o.describeNodes(describe.ToNodeChan(nodes...), o.stream)
	return nil
}

func (o *listPodNodesOptions) listUpOrAll(ctx context.Context) error {
	// filter == all, list all nodes
	// filter == up, list available nodes only
	var ch chan *corepb.Node
	if o.stream { // nolint
		resp, err := o.client.PodNodesStream(ctx, &corepb.ListNodesOptions{
			Podname:         o.name,
			All:             o.filter == all,
			Labels:          o.labels,
			TimeoutInSecond: o.timeoutInSecond,
			SkipInfo:        !o.showInfo,
		})
		if err != nil {
			return err
		}
		ch = make(chan *corepb.Node)
		go func() {
			defer close(ch)
			for {
				node, err := resp.Recv()
				if err != nil {
					if err != io.EOF {
						println(err.Error())
					}
					return
				}
				ch <- node
			}
		}()

	} else {
		resp, err := o.client.ListPodNodes(ctx, &corepb.ListNodesOptions{
			Podname:         o.name,
			All:             o.filter == all,
			Labels:          o.labels,
			TimeoutInSecond: o.timeoutInSecond,
			SkipInfo:        !o.showInfo,
		})
		if err != nil {
			return err
		}
		ch = describe.ToNodeChan(resp.GetNodes()...)

	}

	o.describeNodes(ch, o.stream)
	return nil
}

func (o *listPodNodesOptions) describeNodes(nodes chan *corepb.Node, stream bool) {
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
		stream:          c.Bool("stream"),
	}
	return o.run(c.Context)
}
