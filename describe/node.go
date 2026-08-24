package describe

import (
	"context"
	"encoding/json"
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

// Nodes describes nodes as json, yaml or a table.
func Nodes(nodes <-chan *corepb.Node, stream bool) {
	switch {
	case isJSON():
		describeChAsJSON(nodes)
	case isYAML():
		describeChAsYAML(nodes)
	default:
		describeNodes(nodes, false, stream)
	}
}

// NodesWithInfo describes nodes together with their engine info.
func NodesWithInfo(nodes <-chan *corepb.Node, stream bool) {
	switch {
	case isJSON():
		describeChAsJSON(nodes)
	case isYAML():
		describeChAsYAML(nodes)
	default:
		describeNodes(nodes, true, stream)
	}
}

// NodeResources describes node resource usage as json, yaml or a table.
func NodeResources(ctx context.Context, resources chan *corepb.NodeResource, stream bool) {
	switch {
	case isJSON():
		describeChAsJSON(resources)
	case isYAML():
		describeChAsYAML(resources)
	default:
		describeNodeResources(ctx, resources, stream)
	}
}

// NodeStatusMessage describes node status messages as json, yaml or log lines.
func NodeStatusMessage(ctx context.Context, ms ...*corepb.NodeStatusStreamMessage) {
	switch {
	case isJSON():
		describeAsJSON(ms)
	case isYAML():
		describeAsYAML(ms)
	default:
		describeNodeStatusMessage(ctx, ms)
	}
}

func describeNodes(nodes <-chan *corepb.Node, showInfo, stream bool) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	first := true
	for node := range nodes {
		header, cells := parseNodePluginResources(node)
		if first {
			first = false
			header = append([]any{headerName, "Endpoint", "Status"}, header...)
			if showInfo {
				header = append(header, "Info")
			}
			t.AppendHeader(header)
		}

		status := "DOWN"
		if !node.Bypass && node.Available {
			status = "UP"
		}
		status += fmt.Sprintf("\nbypass %v\navailable %v\ntest %v", node.Bypass, node.Available, node.Test)

		rows := [][]string{
			{node.Name},
			{node.Endpoint},
			{status},
		}
		rows = append(rows, cells...)
		if showInfo {
			rows = append(rows, []string{node.Info})
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

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func parse(key, value any) []string {
	res := []string{}
	if m, ok := value.(map[string]any); ok {
		for k, v := range m {
			res = append(res, fmt.Sprintf("%s[%s]: %v", key, k, toJSON(v)))
		}
	} else if s, ok := value.([]any); ok {
		for i, v := range s {
			res = append(res, fmt.Sprintf("%s[%d]: %v", key, i, toJSON(v)))
		}
	} else {
		res = append(res, fmt.Sprintf("%s: %v", key, toJSON(value)))
	}
	return res
}

func parseNodePluginResources(node *corepb.Node) (header []any, cells [][]string) {
	capacities := resourcetypes.Resources{}
	usages := resourcetypes.Resources{}
	if len(node.ResourceCapacity) > 0 {
		_ = json.Unmarshal([]byte(node.ResourceCapacity), &capacities)
	}
	if len(node.ResourceUsage) > 0 {
		_ = json.Unmarshal([]byte(node.ResourceUsage), &usages)
	}

	for _, plugin := range slices.Sorted(maps.Keys(usages)) {
		header = append(header, plugin)

		row := []string{}
		capacity := capacities[plugin]
		usage := usages[plugin]

		capRows := []string{}
		usageRows := []string{}

		for key, value := range capacity {
			capRows = append(capRows, parse(key, value)...)
			if usage != nil && usage[key] != nil {
				usageRows = append(usageRows, parse(key, usage[key])...)
			}
		}
		row = append(row, "Capacity:")
		row = append(row, capRows...)
		row = append(row, "------------")
		row = append(row, "Usage:")
		row = append(row, usageRows...)
		cells = append(cells, row)
	}
	return header, cells
}

func describeNodeResources(ctx context.Context, resources chan *corepb.NodeResource, stream bool) {
	logger := log.WithFunc("describe.describeNodeResources")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{headerName, "Cpu", "Memory", "Storage", "Volume", "Diffs"})

	for resource := range resources {
		cr, sr, err := ToResourcePrecent(resource)
		if err != nil {
			logger.Error(ctx, err)
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
