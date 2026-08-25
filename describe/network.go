package describe

import (
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func Networks(networks ...*corepb.Network) {
	describeOr(networks, describeNetworks)
}

func describeNetworks(networks []*corepb.Network) {
	nameRow := []string{}
	networkRow := []string{}
	for _, network := range networks {
		nameRow = append(nameRow, network.Name)
		networkRow = append(networkRow, strings.Join(network.Subnets, ","))
	}
	renderTable([]string{headerName, "Network"}, nameRow, networkRow)
}
