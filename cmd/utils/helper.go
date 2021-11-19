package utils

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"

	pb "github.com/projecteru2/core/rpc/gen"
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

// ParseRAMInHuman returns int value in bytes of a human readable string
// e.g. 100KB -> 102400
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

// SplitEquality transfers a list of
// aaa=bbb, xxx=yyy into
// {aaa:bbb, xxx:yyy}
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

// EnvParser .
func EnvParser(b []byte) ([]byte, error) {
	tmpl, err := template.New("tmpl").Option("missingkey=default").Parse(string(b))
	if err != nil {
		return b, err
	}
	out := bytes.Buffer{}
	err = tmpl.Execute(&out, SplitEquality(os.Environ()))
	return out.Bytes(), err
}

// ExitCoder wraps a cli Action function into
// a function with ExitCoder interface
func ExitCoder(f func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := f(c); err != nil {
			return cli.Exit(err, -1)
		}
		return nil
	}
}

// GetResourceOpts generates resource opts from a cli
func GetResourceOpts(c *cli.Context, stringFlags, stringSliceFlags, boolFlags, overrideStringFlags []string) map[string]*pb.RawParam {
	resourceOpts := map[string]*pb.RawParam{}

	for _, flag := range stringFlags {
		resourceOpts[flag] = &pb.RawParam{Value: &pb.RawParam_Str{Str: c.String(flag)}}
	}
	for _, flag := range stringSliceFlags {
		resourceOpts[flag] = &pb.RawParam{Value: &pb.RawParam_StringSlice{StringSlice: &pb.StringSlice{Slice: c.StringSlice(flag)}}} // gross!
	}
	for _, flag := range boolFlags {
		resourceOpts[flag] = nil
	}
	for _, flag := range overrideStringFlags {
		if c.IsSet(flag) {
			resourceOpts[flag] = &pb.RawParam{Value: &pb.RawParam_Str{Str: c.String(flag)}}
		}
	}
	return resourceOpts
}
