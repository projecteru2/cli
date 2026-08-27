package image

import (
	"context"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func resolvePodname(ctx context.Context, client corepb.CoreRPCClient, podname string, nodenames []string) (string, error) {
	if podname != "" || len(nodenames) == 0 {
		return podname, nil
	}
	node, err := client.GetNode(ctx, &corepb.GetNodeOptions{Nodename: nodenames[0]})
	if err != nil {
		return "", err
	}
	return node.Podname, nil
}
