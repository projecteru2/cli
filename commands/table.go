package commands

import (
	"github.com/jedib0t/go-pretty/table"
)

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
