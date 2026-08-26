package node

import (
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v3"
)

func collectResourceParams(cmd *cli.Command) (cpumem, storage resourcetypes.RawParams) {
	cpumem, storage = resourcetypes.RawParams{}, resourcetypes.RawParams{}
	if cmd.IsSet("memory") {
		cpumem["memory"] = cmd.String("memory")
	}
	if cmd.IsSet("numa-cpu") {
		cpumem["numa-cpu"] = cmd.StringSlice("numa-cpu")
	}
	if cmd.IsSet("numa-memory") {
		cpumem["numa-memory"] = cmd.StringSlice("numa-memory")
	}
	if cmd.IsSet("disk") {
		storage["disks"] = cmd.StringSlice("disk")
	}
	if cmd.IsSet(flagStorage) {
		storage[flagStorage] = cmd.String(flagStorage)
	}
	if cmd.IsSet("volume") {
		storage["volumes"] = cmd.StringSlice("volume")
	}
	return cpumem, storage
}
