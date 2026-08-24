package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

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

	describe.Nodes(describe.ToChan(node), false)
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
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func readTLSConfigs(cmd *cli.Command) (caContent, certContent, keyContent string, err error) {
	ca := cmd.String("ca")
	if ca == "" {
		defaultPath := "/etc/docker/tls/ca.crt"
		if _, err := os.Stat(defaultPath); err == nil {
			ca = defaultPath
		}
	}
	if ca != "" {
		f, err := os.ReadFile(ca)
		if err != nil {
			return "", "", "", fmt.Errorf("read %s: %w", ca, err)
		}
		caContent = string(f)
	}

	cert := cmd.String("cert")
	if cert == "" {
		defaultPath := "/etc/docker/tls/client.crt"
		if _, err := os.Stat(defaultPath); err == nil {
			cert = defaultPath
		}
	}
	if cert != "" {
		f, err := os.ReadFile(cert)
		if err != nil {
			return "", "", "", fmt.Errorf("read %s: %w", cert, err)
		}
		certContent = string(f)
	}

	key := cmd.String("key")
	if key == "" {
		defaultPath := "/etc/docker/tls/client.key"
		if _, err := os.Stat(defaultPath); err == nil {
			key = defaultPath
		}
	}
	if key != "" {
		f, err := os.ReadFile(key)
		if err != nil {
			return "", "", "", fmt.Errorf("read %s: %w", key, err)
		}
		keyContent = string(f)
	}
	return caContent, certContent, keyContent, nil
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

	cpumem := resourcetypes.RawParams{}
	storage := resourcetypes.RawParams{}

	if cmd.IsSet("cpu") {
		cpumem["cpu"] = cmd.Int64("cpu")
	}
	if cmd.IsSet("share") {
		cpumem["share"] = cmd.String("share")
	}
	if cmd.IsSet("memory") {
		cpumem["memory"] = cmd.String("memory")
	}
	if cmd.IsSet("numa-cpu") {
		cpumem["numa-cpu"] = cmd.StringSlice("numa-cpu")
	}
	if cmd.IsSet("numa-memory") {
		cpumem["numa-memory"] = cmd.StringSlice("numa-memory")
	}
	if cmd.IsSet("disk") {
		storage["disks"] = cmd.StringSlice("disk")
	}
	if cmd.IsSet("storage") {
		storage["storage"] = cmd.String("storage")
	}
	if cmd.IsSet("volume") {
		storage["volumes"] = cmd.StringSlice("volume")
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)
	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
	}

	if extraResourcesMap, err := utils.ParseExtraResources(cmd); err == nil {
		for k, v := range extraResourcesMap {
			if _, ok := resources[k]; ok {
				continue
			}
			eb, _ := json.Marshal(v)
			resources[k] = eb
		}
	} else {
		return nil, fmt.Errorf("parse extra resources: %w", err)
	}

	labels := utils.SplitEquality(cmd.StringSlice("label"))
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
