package utils

import "github.com/urfave/cli/v3"

// ExtraResourcesFlag is the one --extra-resources every resource-taking command shares.
func ExtraResourcesFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  FlagExtraResources,
		Usage: "extra resource requests as JSON keyed by plugin name, e.g. {\"resource-gpu\":{\"prod_count_map\":{\"nvidia-3070\":1}}}",
	}
}

// FileFlag is the one --file every file-carrying command shares; usage says what the command does with the files.
func FileFlag(usage string) *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name:  FlagFile,
		Usage: usage + ", src_path:dst_path[:mode[:uid:gid]]",
	}
}
