package node

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type addNodeOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.AddNodeOptions
}

func (o *addNodeOptions) run(ctx context.Context) error {
	node, err := o.client.AddNode(ctx, o.opts)
	if err != nil {
		return err
	}

	describe.Nodes(describe.ToChan(node), false, false)
	return nil
}

func cmdNodeAdd(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateAddNodeOptions(cmd)
	if err != nil {
		return err
	}

	o := &addNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(ctx)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

func readTLSConfigs(cmd *cli.Command) (ca, cert, key string, err error) {
	for _, tls := range []struct {
		flag        string
		defaultPath string
		content     *string
	}{
		{flagCA, "/etc/docker/tls/ca.crt", &ca},
		{flagCert, "/etc/docker/tls/client.crt", &cert},
		{flagKey, "/etc/docker/tls/client.key", &key},
	} {
		path := cmd.String(tls.flag)
		if path == "" {
			if _, statErr := os.Stat(tls.defaultPath); statErr != nil {
				continue
			}
			path = tls.defaultPath
		}
		data, readErr := os.ReadFile(path) //nolint:gosec
		if readErr != nil {
			return "", "", "", fmt.Errorf("read %s: %w", path, readErr)
		}
		*tls.content = string(data)
	}
	return ca, cert, key, nil
}

func generateAddNodeOptions(cmd *cli.Command) (*corepb.AddNodeOptions, error) {
	podname := cmd.Args().First()
	if podname == "" {
		return nil, errors.New("podname must not be empty")
	}

	nodename := cmd.String("nodename")

	ca, cert, key, err := readTLSConfigs(cmd)
	if err != nil {
		return nil, err
	}

	endpoint := cmd.String("endpoint")
	if endpoint == "" {
		ip := getLocalIP()
		if ip == "" {
			return nil, errors.New("unable to get local ip")
		}
		port := 2376
		if ca == "" {
			port = 2375
		}
		endpoint = fmt.Sprintf("tcp://%s:%d", ip, port)
	}

	cpumem, storage := collectResourceParams(cmd)
	if cmd.IsSet("cpu") {
		cpumem["cpu"] = cmd.Int("cpu")
	}
	if cmd.IsSet("share") {
		cpumem["share"] = strconv.Itoa(cmd.Int("share"))
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: storage,
	})
	if err != nil {
		return nil, err
	}

	labels := utils.SplitEquality(cmd.StringSlice(flagLabel))
	return &corepb.AddNodeOptions{
		Nodename:  nodename,
		Endpoint:  endpoint,
		Podname:   podname,
		Ca:        ca,
		Cert:      cert,
		Key:       key,
		Labels:    labels,
		Resources: resources,
		Test:      cmd.Bool("test"),
	}, nil
}
