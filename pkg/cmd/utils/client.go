package utils

import (
	coreclient "github.com/projecteru2/core/client"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
)

func NewCoreRPCClient(c *cli.Context) (corepb.CoreRPCClient, error) {
	client, err := coreclient.NewClient(c.Context, c.String("eru"), coretypes.AuthConfig{
		Username: c.String("username"),
		Password: c.String("password"),
	})
	if err != nil {
		return nil, err
	}
	return client.GetRPCClient(), nil
}
