package describe

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

// Core function will describe a coreinfo
// output format can be json or yaml or table
func Core(info *corepb.CoreInfo) {
	switch {
	case isJSON():
		describeAsJSON(info)
	case isYAML():
		describeAsYAML(info)
	default:
		describeCore(info)
	}
}

func describeCore(info *corepb.CoreInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Description"})

	nameRow := []string{"Version", "Git hash", "Built", "Golang version", "OS/Arch", "Identifier"}
	// this stupid Revison typo thing is driving my crazy!!!!!!!!!!!!!!!
	descRow := []string{info.Version, info.Revison, info.BuildAt, info.GolangVersion, info.OsArch, info.Identifier}
	rows := [][]string{nameRow, descRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}
