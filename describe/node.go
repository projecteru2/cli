package describe

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
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
	t.AppendHeader(table.Row{"Name", "Endpoint", "Status", "CPU", "Memory", "Volume", "Storage"})

	for _, node := range nodes {
		totalVolumeCap := int64(0)
		for _, v := range node.InitVolume {
			totalVolumeCap += v
		}

		var status string
		if node.Available {
			status = "UP"
		} else {
			status = "DOWN"
		}
		rows := [][]string{
			{node.Name},
			{node.Endpoint},
			{status},
			{fmt.Sprintf("%.2f / %d", node.CpuUsed, len(node.InitCpu))},
			{fmt.Sprintf("%d / %d bytes", node.MemoryUsed, node.InitMemory)},
			{fmt.Sprintf("%d / %d bytes", node.VolumeUsed, totalVolumeCap)},
			{fmt.Sprintf("%d / %d bytes", node.StorageUsed, node.InitStorage)},
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

// NodeResources describes a list of NodeResource
// output format can be json or yaml or table
func NodeResources(resources ...*corepb.NodeResource) {
	for _, resource := range resources {
		checkNaNForResource(resource)
	}

	switch {
	case isJSON():
		describeAsJSON(resources)
	case isYAML():
		describeAsYAML(resources)
	default:
		describeNodeResources(resources)
	}
}

func checkNaNForResource(resource *corepb.NodeResource) {
	if math.IsNaN(resource.VolumePercent) {
		resource.VolumePercent = 0
	}
	if math.IsNaN(resource.MemoryPercent) {
		resource.MemoryPercent = 0
	}
	if math.IsNaN(resource.StoragePercent) {
		resource.StoragePercent = 0
	}
	if math.IsNaN(resource.CpuPercent) {
		resource.CpuPercent = 0
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
