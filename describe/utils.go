package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

// Format indicates the output format
// can be yaml / yml / json or empty as default
// default will be table
var Format string

func isJSON() bool {
	return strings.ToLower(Format) == "json"
}

func isYAML() bool {
	y := strings.ToLower(Format)
	return y == "yaml" || y == "yml"
}

// actually i need a `zip longest` function
// like in python itertools
func toTableRows(rows [][]string) []table.Row {
	total := len(rows)
	maxLength := 0
	for _, row := range rows {
		if len(row) > maxLength {
			maxLength = len(row)
		}
	}

	rs := []table.Row{}
	for i := 0; i < maxLength; i++ {
		lines := []interface{}{}
		for j := 0; j < total; j++ {
			if i < len(rows[j]) {
				lines = append(lines, rows[j][i])
			} else {
				lines = append(lines, "")
			}
		}
		rs = append(rs, table.Row(lines))
	}
	return rs
}

func describeAsJSON(o interface{}) {
	j, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(j))
}

func describeChAsJSON[T any](ch chan T) {
	for t := range ch {
		j, _ := json.MarshalIndent(t, "", "  ")
		fmt.Println(string(j))
	}
}

func describeAsYAML(o interface{}) {
	y, _ := yaml.Marshal(o)
	fmt.Println(string(y))
}

func describeChAsYAML[T any](ch chan T) {
	for t := range ch {
		j, _ := yaml.Marshal(t)
		fmt.Println(string(j))
	}
}

// ToNodeChan is to be rewritten using generic
func ToNodeChan(nodes ...*corepb.Node) chan *corepb.Node {
	ch := make(chan *corepb.Node)
	go func() {
		defer close(ch)
		for _, node := range nodes {
			ch <- node
		}
	}()
	return ch
}

func ToNodeResourceChan(resources ...*corepb.NodeResource) chan *corepb.NodeResource {
	ch := make(chan *corepb.NodeResource)
	go func() {
		defer close(ch)
		for _, resource := range resources {
			checkNaNForResource(resource)
			ch <- resource
		}
	}()
	return ch
}
