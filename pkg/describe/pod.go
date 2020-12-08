package describe

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

func DescribePods(pods ...*corepb.Pod) {
	switch strings.ToLower(Format) {
	case "json":
		describeAsJSON(pods)
	case "yaml", "yml":
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
