package pod

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

var filterExpr = regexp.MustCompile(`^\s*(?P<name>cpu|memory|storage|volume)\s*(?P<op>>=|<=|==|>|<)\s*(?P<value>\d+(?:\.\d+)?%?)\s*$`)

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

func attr(nr *corepb.NodeResource, name string) float64 {
	cr, sr, err := describe.ToResourcePrecent(nr)
	if err != nil {
		return 0.0
	}
	switch name {
	case flagCPU:
		return cr[flagCPU]
	case flagMemory:
		return cr[flagMemory]
	case flagStorage:
		return sr[flagStorage]
	case "volume":
		return sr["volumes"]
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
	if o.expr == "" {
		return ch, nil
	}

	filter := match(o.expr)
	if len(filter) == 0 {
		return nil, fmt.Errorf("invalid filter %q, want one of cpu/memory/storage/volume with an operator and a value", o.expr)
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
			if !compare(filter["op"], l, v) {
				continue
			}
			rv <- nr
		}
	}()
	return rv, nil
}

func (o *resourcePodOptions) run(ctx context.Context) error {
	resp, err := o.client.GetPodResource(ctx, &corepb.GetPodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	var recvErr error
	ch := make(chan *corepb.NodeResource)
	go func() {
		defer close(ch)
		for {
			resource, streamErr := resp.Recv()
			if streamErr != nil {
				if !errors.Is(streamErr, io.EOF) {
					recvErr = streamErr
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

	describe.NodeResources(ctx, resChan, o.stream)
	return recvErr
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

	o := &resourcePodOptions{
		client: client,
		name:   name,
		expr:   cmd.String("filter"),
		stream: cmd.Bool("stream"),
	}
	return o.run(ctx)
}
