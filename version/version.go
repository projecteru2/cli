package version

import (
	"fmt"
	"runtime"
)

var (
	NAME     = "eru-cli"
	VERSION  = "unknown"
	REVISION = "HEAD"
	BUILTAT  = "now"
)

// String renders the build identity.
func String() string {
	return fmt.Sprintf(
		"Version:        %s\nGit hash:       %s\nBuilt:          %s\nGolang version: %s\nOS/Arch:        %s/%s\n",
		VERSION, REVISION, BUILTAT, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}
