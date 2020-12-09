package describe

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

// Nodes describes a list of Node
// output format can be json or yaml or table
func Nodes(nodes ...*corepb.Node) {
	switch {
	case isJSON():
		describeAsJSON(nodes)
	case isYAML():
		describeAsYAML(nodes)
	default:
		describeNodes(nodes)
	}
}

func describeNodes(nodes []*corepb.Node) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Endpoint"})

	nameRow := []string{}
	endpointRow := []string{}
	for _, node := range nodes {
		nameRow = append(nameRow, node.Name)
		endpointRow = append(endpointRow, node.Endpoint)
	}
	rows := [][]string{nameRow, endpointRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}

// NodeResources describes a list of NodeResource
// output format can be json or yaml or table
func NodeResources(resources ...*corepb.NodeResource) {
	switch {
	case isJSON():
		describeAsJSON(resources)
	case isYAML():
		describeAsYAML(resources)
	default:
		describeNodeResources(resources)
	}
}

func describeNodeResources(resources []*corepb.NodeResource) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Cpu", "Memory", "Storage", "Volume", "Diffs"})

	nodeRow := []string{}
	cpuRow := []string{}
	memoryRow := []string{}
	storageRow := []string{}
	volumeRow := []string{}
	diffsRow := []string{}
	for _, resource := range resources {
		nodeRow = append(nodeRow, resource.Name)
		cpuRow = append(cpuRow, fmt.Sprintf("%.2f%%", resource.CpuPercent*100))
		memoryRow = append(memoryRow, fmt.Sprintf("%.2f%%", resource.MemoryPercent*100))
		storageRow = append(storageRow, fmt.Sprintf("%.2f%%", resource.StoragePercent*100))
		volumeRow = append(volumeRow, fmt.Sprintf("%.2f%%", resource.VolumePercent*100))
		diffsRow = append(diffsRow, strings.Join(resource.Diffs, "\n"))
	}
	rows := [][]string{nodeRow, cpuRow, memoryRow, storageRow, volumeRow, diffsRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}
