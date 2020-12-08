package describe

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

func DescribeNetworks(networks ...*corepb.Network) {
	switch strings.ToLower(Format) {
	case "json":
		describeAsJSON(networks)
	case "yaml", "yml":
		describeAsYAML(networks)
	default:
		describeNetworks(networks)
	}
}

func describeNetworks(networks []*corepb.Network) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Network"})

	nameRow := []string{}
	networkRow := []string{}
	for _, network := range networks {
		nameRow = append(nameRow, network.Name)
		networkRow = append(networkRow, strings.Join(network.GetSubnets(), ","))
	}
	rows := [][]string{nameRow, networkRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}
