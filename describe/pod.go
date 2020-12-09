package describe

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

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
