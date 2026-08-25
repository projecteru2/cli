package utils

import (
	"context"

	coreclient "github.com/projecteru2/core/client"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"
)

// NewCoreRPCClient dials the core address in cmd and returns its RPC client.
func NewCoreRPCClient(ctx context.Context, cmd *cli.Command) (corepb.CoreRPCClient, error) {
	client, err := coreclient.NewClient(ctx, cmd.String("eru"), coretypes.AuthConfig{
		Username: cmd.String("username"),
		Password: cmd.String("password"),
	})
	if err != nil {
		return nil, err
	}
	return client.GetRPCClient(), nil
}
