package pod

import (
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/juju/errors"
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
	// TODO:
	// re-implements after the proto is ready.
	// case name == "cpu":
	// 	return nr.CpuPercent
	// case name == "memory":
	// 	return nr.MemoryPercent
	// case name == "storage":
	// 	return nr.StoragePercent
	// case name == "volume":
	// 	return nr.VolumePercent
	default:
		return 0
	}
}

type resourcePodOptions struct {
	client corepb.CoreRPCClient
	name   string
	expr   string
	stream bool
}

func (o *resourcePodOptions) filter(ch chan *corepb.NodeResource) (chan *corepb.NodeResource, error) {
	filter := match(o.expr)
	if len(filter) == 0 {
		return ch, nil
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

	rv := make(chan *corepb.NodeResource)
	go func() {
		defer close(rv)
		for nr := range ch {
			l := attr(nr, filter["name"])
			if !op(filter["op"], l, v) {
				continue
			}
			rv <- nr
		}
	}()
	return rv, nil
}

func (o *resourcePodOptions) run(ctx context.Context) error {
	var ch chan *corepb.NodeResource
	resp, err := o.client.GetPodResource(ctx, &corepb.GetPodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	ch = make(chan *corepb.NodeResource)
	go func() {
		defer close(ch)
		for {
			resource, err := resp.Recv()
			if err != nil {
				if err != io.EOF {
					println(err.Error())
				}
				return
			}
			ch <- resource
		}
	}()

	resChan, err := o.filter(ch)
	if err != nil {
		return err
	}

	describe.NodeResources(resChan, o.stream)
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
		stream: c.Bool("stream"),
	}
	return o.run(c.Context)
}
