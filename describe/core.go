package describe

import (
	corepb "github.com/projecteru2/core/rpc/gen"
)

func Core(info *corepb.CoreInfo) {
	describeOr(info, describeCore)
}

func describeCore(info *corepb.CoreInfo) {
	names := []string{"Version", "Git hash", "Built", "Golang version", "OS/Arch", "Identifier"}
	// Revison is misspelled in the core protobuf definition.
	values := []string{info.Version, info.Revison, info.BuildAt, info.GolangVersion, info.OsArch, info.Identifier}
	renderTable([]string{headerName, "Description"}, [][]string{names, values})
}
