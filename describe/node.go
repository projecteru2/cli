package describe

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
)

// Nodes describes nodes a command already holds.
func Nodes(showInfo bool, nodes ...*corepb.Node) {
	describeOr(nodes, func(all []*corepb.Node) { renderNodes(showInfo, all...) })
}

// NodesStream describes nodes as they arrive, one table per node when stream is set.
func NodesStream(nodes <-chan *corepb.Node, showInfo, stream bool) {
	describeChOr(nodes, func(ch <-chan *corepb.Node) { describeNodes(ch, showInfo, stream) })
}

// NodeResource describes one node's resource.
func NodeResource(ctx context.Context, resource *corepb.NodeResource) {
	describeOr(resource, func(r *corepb.NodeResource) { renderNodeResources(ctx, r) })
}

func NodeResources(ctx context.Context, resources <-chan *corepb.NodeResource, stream bool) {
	describeChOr(resources, func(ch <-chan *corepb.NodeResource) { describeNodeResources(ctx, ch, stream) })
}

// NodeStatusMessage describes node status messages as json, yaml or log lines.
func NodeStatusMessage(ctx context.Context, ms ...*corepb.NodeStatusStreamMessage) {
	describeOr(ms, func(m []*corepb.NodeStatusStreamMessage) { describeNodeStatusMessage(ctx, m) })
}

func describeNodes(nodes <-chan *corepb.Node, showInfo, stream bool) {
	if stream {
		for node := range nodes {
			renderNodes(showInfo, node)
		}
		return
	}

	all := []*corepb.Node{}
	for node := range nodes {
		all = append(all, node)
	}
	renderNodes(showInfo, all...)
}

func renderNodes(showInfo bool, nodes ...*corepb.Node) {
	capacities := make([]resourcetypes.Resources, len(nodes))
	usages := make([]resourcetypes.Resources, len(nodes))
	for i, node := range nodes {
		capacities[i] = unmarshalResources(node.ResourceCapacity)
		usages[i] = unmarshalResources(node.ResourceUsage)
	}
	names := pluginNames(capacities, usages)

	header := append([]string{headerName, "Endpoint", "Status"}, names...)
	if showInfo {
		header = append(header, "Info")
	}

	groups := make([][][]string, 0, len(nodes))
	for i, node := range nodes {
		status := "DOWN"
		if !node.Bypass && node.Available {
			status = "UP"
		}
		status += fmt.Sprintf("\nbypass %v\navailable %v\ntest %v", node.Bypass, node.Available, node.Test)

		rows := [][]string{{node.Name}, {node.Endpoint}, {status}}
		for _, name := range names {
			rows = append(rows, nodePluginRows(capacities[i][name], usages[i][name]))
		}
		if showInfo {
			rows = append(rows, []string{node.Info})
		}
		groups = append(groups, rows)
	}

	renderTable(header, groups...)
}

func nodePluginRows(capacity, usage resourcetypes.RawParams) []string {
	if len(capacity) == 0 && len(usage) == 0 {
		return nil
	}
	rows := []string{"Capacity:"}
	rows = append(rows, parseAll(capacity)...)
	rows = append(rows, "------------", "Usage:")
	return append(rows, parseAll(usage)...)
}

func describeNodeResources(ctx context.Context, resources <-chan *corepb.NodeResource, stream bool) {
	if stream {
		for resource := range resources {
			renderNodeResources(ctx, resource)
		}
		return
	}
	all := []*corepb.NodeResource{}
	for resource := range resources {
		all = append(all, resource)
	}
	renderNodeResources(ctx, all...)
}

func renderNodeResources(ctx context.Context, resources ...*corepb.NodeResource) {
	logger := log.WithFunc("describe.renderNodeResources")
	groups := make([][][]string, 0, len(resources))
	for _, resource := range resources {
		cr, sr, err := ToResourcePercent(resource)
		if err != nil {
			logger.Errorf(ctx, err, "resource percent of node %s", resource.Name)
			continue
		}
		groups = append(groups, [][]string{
			{resource.Name},
			{fmt.Sprintf("%.2f%%", cr["cpu"]*100)},
			{fmt.Sprintf("%.2f%%", cr["memory"]*100)},
			{fmt.Sprintf("%.2f%%", sr["storage"]*100)},
			{fmt.Sprintf("%.2f%%", sr["volumes"]*100)},
			{strings.Join(resource.Diffs, "\n")},
		})
	}
	renderTable([]string{headerName, "Cpu", "Memory", "Storage", "Volume", "Diffs"}, groups...)
}

func describeNodeStatusMessage(ctx context.Context, ms []*corepb.NodeStatusStreamMessage) {
	logger := log.WithFunc("describe.describeNodeStatusMessage")
	for _, m := range ms {
		if m.Error != "" {
			logger.Errorf(ctx, errors.New(m.Error), "get status for node %s", m.Nodename)
		} else {
			logger.Infof(ctx, "node %s on pod %s, alive: %v", m.Nodename, m.Podname, m.Alive)
		}
	}
}
