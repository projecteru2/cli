package pod

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

var filterExpr = regexp.MustCompile(`^\s*(?P<name>cpu|memory|storage|volume)\s*(?P<op>>=|<=|==|>|<)\s*(?P<value>\d+(?:\.\d+)?%?)\s*$`)

type resourcePodOptions struct {
	client corepb.CoreRPCClient
	name   string
	keep   describe.NodeResourceFilter
	stream bool
}

func (o *resourcePodOptions) run(ctx context.Context) error {
	resp, err := o.client.GetPodResource(ctx, &corepb.GetPodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	ch, wait := utils.StreamToChan(resp.Recv)
	describe.NodeResources(ctx, ch, o.stream, o.keep)
	return wait()
}

func cmdPodResource(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("pod name must be given")
	}

	keep, err := parseFilter(cmd.String("filter"))
	if err != nil {
		return err
	}

	o := &resourcePodOptions{
		client: client,
		name:   name,
		keep:   keep,
		stream: cmd.Bool("stream"),
	}
	return o.run(ctx)
}

func parseFilter(expr string) (describe.NodeResourceFilter, error) {
	if expr == "" {
		return nil, nil
	}

	filter := match(expr)
	if len(filter) == 0 {
		return nil, fmt.Errorf("invalid filter %q, want one of cpu/memory/storage/volume with an operator and a value", expr)
	}

	value, percent := strings.CutSuffix(filter["value"], "%")
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	if percent {
		v /= 100
	}

	name, op := filter["name"], filter["op"]
	return func(cpumem, storage map[string]float64) bool {
		return compare(op, attr(cpumem, storage, name), v)
	}, nil
}

func match(s string) map[string]string {
	rv := make(map[string]string)
	founds := filterExpr.FindStringSubmatch(s)
	for i, name := range filterExpr.SubexpNames() {
		if i > 0 && i < len(founds) {
			rv[name] = founds[i]
		}
	}
	return rv
}

func compare(operator string, left, right float64) bool {
	switch operator {
	case ">":
		return left > right
	case ">=":
		return left >= right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case "==":
		return left == right
	default:
		return false
	}
}

func attr(cpumem, storage map[string]float64, name string) float64 {
	switch name {
	case flagCPU:
		return cpumem[flagCPU]
	case flagMemory:
		return cpumem[flagMemory]
	case flagStorage:
		return storage[flagStorage]
	case "volume":
		return storage["volumes"]
	default:
		return 0
	}
}
