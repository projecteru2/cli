package describe

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

type capacityOfNode struct {
	Name     string `json:"name" yaml:"name"`
	Capacity int    `json:"capacity" yaml:"capacity"`
}

type capacityOfPod struct {
	Total int               `json:"total" yaml:"total"`
	Nodes []*capacityOfNode `json:"nodes" yaml:"nodes"`
}

// Pods describes a list of Pod
// output format can be json or yaml or table
func Pods(pods ...*corepb.Pod) {
	switch {
	case isJSON():
		describeAsJSON(pods)
	case isYAML():
		describeAsYAML(pods)
	default:
		describePods(pods)
	}
}

// PodCapacity describes the capacity remained based on a given specification.
// output format can be json or yaml or table
func PodCapacity(total int64, capacityMap map[string]int64) {
	capPod := &capacityOfPod{
		Total: int(total),
	}

	capPod.Nodes = make([]*capacityOfNode, 0, len(capacityMap))
	for name, capacity := range capacityMap {
		capPod.Nodes = append(capPod.Nodes, &capacityOfNode{
			Name:     name,
			Capacity: int(capacity),
		})
	}

	// sort by remained capacity in descending order
	sort.Slice(capPod.Nodes, func(i, j int) bool {
		return capPod.Nodes[i].Capacity >= capPod.Nodes[j].Capacity
	})

	switch {
	case isJSON():
		describeAsJSON(capPod)
	case isYAML():
		describeAsYAML(capPod)
	default:
		describePodCapacities(capPod)
	}
}

func describePodCapacities(capacity *capacityOfPod) {
	fmt.Println("Total:", capacity.Total)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Node", "Capacity"})

	nameRow := []string{}
	descRow := []string{}
	for _, node := range capacity.Nodes {
		nameRow = append(nameRow, node.Name)
		descRow = append(descRow, strconv.FormatInt(int64(node.Capacity), 10))
	}
	rows := [][]string{nameRow, descRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}

func describePods(pods []*corepb.Pod) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Description"})

	nameRow := []string{}
	descRow := []string{}
	for _, pod := range pods {
		nameRow = append(nameRow, pod.Name)
		descRow = append(descRow, pod.Desc)

	}
	rows := [][]string{nameRow, descRow}
	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}
