package utils

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"text/template"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v3"
)

// GetNetworks returns a networkmode -> ip map
func GetNetworks(network string) map[string]string {
	var ip string
	networkInfo := strings.Split(network, "=")
	if len(networkInfo) == 2 {
		network = networkInfo[0]
		ip = networkInfo[1]
	}
	networks := map[string]string{}
	if network != "" {
		networks[network] = ip
	}
	return networks
}

// ParseRAMInHuman converts a human readable size such as 100KB into bytes.
func ParseRAMInHuman(ram string) (int64, error) {
	if ram == "" {
		return 0, nil
	}
	flag := int64(1)
	if strings.HasPrefix(ram, "-") {
		flag = int64(-1)
		ram = strings.TrimLeft(ram, "-")
	}
	ramInBytes, err := units.RAMInBytes(ram)
	if err != nil {
		return 0, err
	}
	return ramInBytes * flag, nil
}

// SplitEquality turns a list of key=value strings into a map.
func SplitEquality(elements []string) map[string]string {
	r := map[string]string{}
	for _, e := range elements {
		p := strings.SplitN(e, "=", 2)
		if len(p) != 2 {
			continue
		}
		r[p[0]] = p[1]
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

func GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
