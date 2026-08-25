package describe

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

type capacityOfNode struct {
	Name     string `json:"name" yaml:"name"`
	Capacity int64  `json:"capacity" yaml:"capacity"`
}

type capacityOfPod struct {
	Total int64             `json:"total" yaml:"total"`
	Nodes []*capacityOfNode `json:"nodes" yaml:"nodes"`
}

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

// PodCapacity describes the capacity left for a given specification.
func PodCapacity(total int64, capacityMap map[string]int64) {
	capPod := &capacityOfPod{Total: total}

	capPod.Nodes = make([]*capacityOfNode, 0, len(capacityMap))
	for name, capacity := range capacityMap {
		capPod.Nodes = append(capPod.Nodes, &capacityOfNode{
			Name:     name,
			Capacity: capacity,
		})
	}

	slices.SortFunc(capPod.Nodes, func(a, b *capacityOfNode) int {
		return cmp.Or(cmp.Compare(b.Capacity, a.Capacity), cmp.Compare(a.Name, b.Name))
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

func describePods(pods []*corepb.Pod) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{headerName, "Description"})

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

func describePodCapacities(capacity *capacityOfPod) {
	fmt.Println("Total:", capacity.Total)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Node", "Capacity"})

	nameRow := []string{}
	descRow := []string{}
	for _, node := range capacity.Nodes {
		nameRow = append(nameRow, node.Name)
		descRow = append(descRow, strconv.FormatInt(node.Capacity, 10))
	}
	rows := [][]string{nameRow, descRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}
