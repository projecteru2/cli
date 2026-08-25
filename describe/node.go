package describe

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
)

func Nodes(nodes <-chan *corepb.Node, showInfo, stream bool) {
	describeChOr(nodes, func(ch <-chan *corepb.Node) { describeNodes(ch, showInfo, stream) })
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
	if len(nodes) == 0 {
		return
	}

	capacities := make([]resourcetypes.Resources, len(nodes))
	usages := make([]resourcetypes.Resources, len(nodes))
	plugins := map[string]struct{}{}
	for i, node := range nodes {
		capacities[i] = unmarshalResources(node.ResourceCapacity)
		usages[i] = unmarshalResources(node.ResourceUsage)
		for plugin := range capacities[i] {
			plugins[plugin] = struct{}{}
		}
		for plugin := range usages[i] {
			plugins[plugin] = struct{}{}
		}
	}
	names := slices.Sorted(maps.Keys(plugins))

	header := []any{headerName, "Endpoint", "Status"}
	for _, name := range names {
		header = append(header, name)
	}
	if showInfo {
		header = append(header, "Info")
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)

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
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
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
	logger := log.WithFunc("describe.describeNodeResources")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{headerName, "Cpu", "Memory", "Storage", "Volume", "Diffs"})

	for resource := range resources {
		cr, sr, err := ToResourcePercent(resource)
		if err != nil {
			logger.Errorf(ctx, err, "resource percent of node %s", resource.Name)
			continue
		}
		rows := [][]string{
			{resource.Name},
			{fmt.Sprintf("%.2f%%", cr["cpu"]*100)},
			{fmt.Sprintf("%.2f%%", cr["memory"]*100)},
			{fmt.Sprintf("%.2f%%", sr["storage"]*100)},
			{fmt.Sprintf("%.2f%%", sr["volumes"]*100)},
			{strings.Join(resource.Diffs, "\n")},
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
		if stream {
			t.SetStyle(table.StyleLight)
			t.Render()
			t.ResetRows()
		}
	}
	if !stream {
		t.SetStyle(table.StyleLight)
		t.Render()
	}
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
