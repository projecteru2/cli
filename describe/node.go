package describe

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/jedib0t/go-pretty/v6/table"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/sirupsen/logrus"
)

// Nodes describes a list of Node
// output format can be json or yaml or table
func Nodes(nodes <-chan *corepb.Node, stream bool) {
	switch {
	case isJSON():
		describeChNodeAsJSON(nodes)
	case isYAML():
		describeChNodeAsYAML(nodes)
	default:
		describeNodes(nodes, false, stream)
	}
}

// NodesWithInfo describes a list of Node with their info
func NodesWithInfo(nodes <-chan *corepb.Node, stream bool) {
	switch {
	case isJSON():
		describeChNodeAsJSON(nodes)
	case isYAML():
		describeChNodeAsYAML(nodes)
	default:
		describeNodes(nodes, true, stream)
	}
}

func describeNodes(nodes <-chan *corepb.Node, showInfo, stream bool) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	var once sync.Once

	for node := range nodes {
		header, cells := parseNodePluginResources(node)
		once.Do(func() {
			header = append([]interface{}{"Name", "Endpoint", "Status"}, header...)
			if showInfo {
				header = append(header, "Info")
			}
			t.AppendHeader(header)
		})

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

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func parse(key, value interface{}) []string {
	res := []string{}
	if m, ok := value.(map[string]interface{}); ok {
		for k, v := range m {
			res = append(res, fmt.Sprintf("%s[%s]: %v", key, k, toJSON(v)))
		}
	} else if s, ok := value.([]interface{}); ok {
		for i, v := range s {
			res = append(res, fmt.Sprintf("%s[%d]: %v", key, i, toJSON(v)))
		}
	} else {
		res = append(res, fmt.Sprintf("%s: %v", key, toJSON(value)))
	}
	return res
}

func parseNodePluginResources(node *corepb.Node) (header []interface{}, cells [][]string) {
	capacities := resourcetypes.Resources{}
	usages := resourcetypes.Resources{}
	if len(node.ResourceCapacity) > 0 {
		_ = json.Unmarshal([]byte(node.ResourceCapacity), &capacities)
	}
	if len(node.ResourceUsage) > 0 {
		_ = json.Unmarshal([]byte(node.ResourceUsage), &usages)
	}

	for plugin := range usages {
		header = append(header, plugin)
	}
	sort.Slice(header, func(i, j int) bool {
		return header[i].(string) < header[j].(string)
	})

	for _, plugin := range header {
		row := []string{}
		capacity := capacities[plugin.(string)]
		usage := usages[plugin.(string)]

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

// NodeResources describes a list of NodeResource
// output format can be json or yaml or table
func NodeResources(resources chan *corepb.NodeResource, stream bool) {
	switch {
	case isJSON():
		describeChNodeResourceAsJSON(resources)
	case isYAML():
		describeChNodeResourceAsYAML(resources)
	default:
		describeNodeResources(resources, stream)
	}
}

func describeNodeResources(resources chan *corepb.NodeResource, stream bool) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Cpu", "Memory", "Storage", "Volume", "Diffs"})

	for resource := range resources {
		cr, sr, err := ToResourcePrecent(resource)
		if err != nil {
			logrus.Error(err)
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

// NodeStatusMessage describes NodeStatusStreamMessage
// in json / yaml, or just a line in stdout
func NodeStatusMessage(ms ...*corepb.NodeStatusStreamMessage) {
	switch {
	case isJSON():
		describeAsJSON(ms)
	case isYAML():
		describeAsYAML(ms)
	default:
		describeNodeStatusMessage(ms)
	}
}

func describeNodeStatusMessage(ms []*corepb.NodeStatusStreamMessage) {
	for _, m := range ms {
		if m.Error != "" {
			logrus.Errorf("[WatchNodeStatus] Error when get status for node %s: %s", m.Nodename, m.Error)
		} else {
			logrus.Infof("[WatchNodeStatus] Node %s on pod %s, alive: %v", m.Nodename, m.Podname, m.Alive)
		}
	}
}
