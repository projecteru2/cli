package pod

import (
	"context"
	"encoding/json"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type listPodNodesOptions struct {
	client corepb.CoreRPCClient
	name   string
	all    bool
}

func (o *listPodNodesOptions) run(ctx context.Context) error {
	resp, err := o.client.ListPodNodes(ctx, &corepb.ListNodesOptions{
		Podname: o.name,
		All:     o.all,
	})
	if err != nil {
		return err
	}

	for _, node := range resp.GetNodes() {
		logrus.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
		r := map[string]interface{}{}
		if err := json.Unmarshal([]byte(node.Info), &r); err != nil {
			logrus.Errorf("Get Node Info failed: %v", node.Info)
		}
	}
	return nil
}

func cmdPodListNodes(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &listPodNodesOptions{
		client: client,
		name:   name,
		all:    c.Bool("all"),
	}
	return o.run(c.Context)
}
