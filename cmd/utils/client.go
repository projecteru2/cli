package utils

import (
	"context"

	coreclient "github.com/projecteru2/core/client"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"
)

// NewCoreRPCClient returns an RPC client to use
// it actually wraps the GetRPCClient method
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
