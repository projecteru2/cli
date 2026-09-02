package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"text/template"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
)

// GetNetworks returns a networkmode -> ip map.
func GetNetworks(network string) map[string]string {
	var ip string
	if name, address, ok := strings.Cut(network, "="); ok && !strings.Contains(address, "=") {
		network, ip = name, address
	}
	networks := map[string]string{}
	if network != "" {
		networks[network] = ip
	}
	return networks
}

// ParseRAMInHuman parses a human-readable size ("100KB", "-1T") into bytes; the implementation lives with core's RawParams.
func ParseRAMInHuman(ram string) (int64, error) {
	return resourcetypes.ParseRAMInHuman(ram)
}

// ParseDeployStrategy maps a --deploy-strategy value onto the core enum.
func ParseDeployStrategy(name string) (corepb.DeployOptions_Strategy, error) {
	value, ok := corepb.DeployOptions_Strategy_value[strings.ToUpper(name)]
	if !ok {
		return 0, fmt.Errorf("invalid deploy strategy %q, want one of %s", name, strings.Join(slices.Sorted(maps.Keys(corepb.DeployOptions_Strategy_value)), "/"))
	}
	return corepb.DeployOptions_Strategy(value), nil
}

// SplitEquality turns a list of key=value strings into a map.
func SplitEquality(elements []string) map[string]string {
	r := map[string]string{}
	for _, e := range elements {
		if k, v, ok := strings.Cut(e, "="); ok {
			r[k] = v
		}
	}
	return r
}

// EnvParser expands Go template references to environment variables in b.
func EnvParser(b []byte) ([]byte, error) {
	tmpl, err := template.New("tmpl").Option("missingkey=default").Parse(string(b))
	if err != nil {
		return b, err
	}
	out := bytes.Buffer{}
	err = tmpl.Execute(&out, SplitEquality(os.Environ()))
	return out.Bytes(), err
}

// ExitCoder turns an action error into a cli.ExitCoder.
func ExitCoder(f cli.ActionFunc) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		if err := f(ctx, cmd); err != nil {
			if exitErr, ok := errors.AsType[cli.ExitCoder](err); ok {
				return cli.Exit(exitErr, exitErr.ExitCode())
			}
			return cli.Exit(err, -1)
		}
		return nil
	}
}

// GetHostname returns the local hostname, empty when it cannot be read.
func GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
