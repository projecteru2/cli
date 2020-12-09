package describe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jedib0t/go-pretty/v6/table"
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

func describeAsYAML(o interface{}) {
	y, _ := yaml.Marshal(o)
	fmt.Println(string(y))
}
