package describe

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	corepb "github.com/projecteru2/core/rpc/gen"
)

func DescribeCore(info *corepb.CoreInfo) {
	switch strings.ToLower(Format) {
	case "json":
		describeAsJSON(info)
	case "yaml", "yml":
		describeAsYAML(info)
	default:
		describeCore(info)
	}
}

func describeCore(info *corepb.CoreInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Description"})

	nameRow := []string{"Version", "Git hash", "Built", "Golang version", "OS/Arch"}
	// this stupid Revison typo thing is driving my crazy!!!!!!!!!!!!!!!
	descRow := []string{info.Version, info.Revison, info.BuildAt, info.GolangVersion, info.OsArch}
	rows := [][]string{nameRow, descRow}

	t.AppendRows(toTableRows(rows))
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}
