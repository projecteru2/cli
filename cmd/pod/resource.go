package pod

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

var re = regexp.MustCompile(`(?P<name>cpu|memory|storage|volume)\s*(?P<op>>|>=|<|<=|==)\s*(?P<value>\d+.?\d*%?)`)

func match(s string) map[string]string {
	rv := make(map[string]string)
	founds := re.FindStringSubmatch(s)
	for i, name := range re.SubexpNames() {
		if i > 0 && i < len(founds) {
			rv[name] = founds[i]
		}
	}
	return rv
}

func op(op string, left, right float64) bool {
	switch {
	case op == ">":
		return left > right
	case op == ">=":
		return left >= right
	case op == "<":
		return left < right
	case op == "<=":
		return left <= right
	case op == "==":
		return left == right
	default:
		return false
	}
}

func attr(nr *corepb.NodeResource, name string) float64 {
	switch {
	case name == "cpu":
		return nr.CpuPercent
	case name == "memory":
		return nr.MemoryPercent
	case name == "storage":
		return nr.StoragePercent
	case name == "volume":
		return nr.VolumePercent
	default:
		return 0
	}
}

type resourcePodOptions struct {
	client corepb.CoreRPCClient
	name   string
	expr   string
}

func (o *resourcePodOptions) filter(nrs []*corepb.NodeResource) ([]*corepb.NodeResource, error) {
	filter := match(o.expr)
	if len(filter) == 0 {
		return nrs, nil
	}

	var (
		value   = filter["value"]
		percent bool
	)
	if strings.HasSuffix(value, "%") {
		value = value[:len(value)-1]
		percent = true
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	if percent {
		v /= 100
	}

	rv := []*corepb.NodeResource{}
	for _, nr := range nrs {
		l := attr(nr, filter["name"])
		if !op(filter["op"], l, v) {
			continue
		}
		rv = append(rv, nr)
	}
	return rv, nil
}

func (o *resourcePodOptions) run(ctx context.Context) error {
	resp, err := o.client.GetPodResource(ctx, &corepb.GetPodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	nrs, err := o.filter(resp.NodesResource)
	if err != nil {
		return err
	}

	describe.NodeResources(nrs...)
	return nil
}

func cmdPodResource(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &resourcePodOptions{
		client: client,
		name:   name,
		expr:   c.String("filter"),
	}
	return o.run(c.Context)
}
