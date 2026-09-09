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

type NodeResourceFilter func(cpumem, storage map[string]float64) bool

// Nodes describes nodes a command already holds.
func Nodes(showInfo bool, nodes ...*corepb.Node) {
	describeOr(nodes, func(all []*corepb.Node) { renderNodes(showInfo, all...) })
}

// NodesStream describes nodes as they arrive, one table per node when stream is set.
func NodesStream(nodes <-chan *corepb.Node, showInfo, stream bool) {
	describeChOr(nodes, stream, func(all ...*corepb.Node) { renderNodes(showInfo, all...) })
}

// NodeResource describes one node's resource.
func NodeResource(ctx context.Context, resource *corepb.NodeResource) {
	describeOr(resource, func(r *corepb.NodeResource) { renderNodeResources(nodePercents(ctx, r)...) })
}

func NodeResources(ctx context.Context, resources <-chan *corepb.NodeResource, stream bool, keep NodeResourceFilter) {
	describeChOr(nodePercentChan(ctx, resources, keep), stream, renderNodeResources)
}

// NodeStatusMessage describes node status messages as json, yaml or log lines.
func NodeStatusMessage(ctx context.Context, ms ...*corepb.NodeStatusStreamMessage) {
	describeOr(ms, func(m []*corepb.NodeStatusStreamMessage) { describeNodeStatusMessage(ctx, m) })
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

type nodePercent struct {
	*corepb.NodeResource
	cpumem  map[string]float64
	storage map[string]float64
}

func nodePercents(ctx context.Context, resources ...*corepb.NodeResource) []nodePercent {
	rv := make([]nodePercent, 0, len(resources))
	for _, resource := range resources {
		if percent, ok := nodePercentOf(ctx, resource); ok {
			rv = append(rv, percent)
		}
	}
	return rv
}

func nodePercentOf(ctx context.Context, resource *corepb.NodeResource) (nodePercent, bool) {
	cr, sr, err := ToResourcePercent(resource)
	if err != nil {
		log.WithFunc("describe.nodePercentOf").Errorf(ctx, err, "resource percent of node %s", resource.Name)
		return nodePercent{}, false
	}
	return nodePercent{resource, cr, sr}, true
}

func nodePercentChan(ctx context.Context, resources <-chan *corepb.NodeResource, keep NodeResourceFilter) <-chan nodePercent {
	rv := make(chan nodePercent)
	go func() {
		defer close(rv)
		for resource := range resources {
			if percent, ok := nodePercentOf(ctx, resource); ok && (keep == nil || keep(percent.cpumem, percent.storage)) {
				rv <- percent
			}
		}
	}()
	return rv
}

func renderNodeResources(resources ...nodePercent) {
	groups := make([][][]string, 0, len(resources))
	for _, resource := range resources {
		groups = append(groups, [][]string{
			{resource.Name},
			{fmt.Sprintf("%.2f%%", resource.cpumem["cpu"]*100)},
			{fmt.Sprintf("%.2f%%", resource.cpumem["memory"]*100)},
			{fmt.Sprintf("%.2f%%", resource.storage["storage"]*100)},
			{fmt.Sprintf("%.2f%%", resource.storage["volumes"]*100)},
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
