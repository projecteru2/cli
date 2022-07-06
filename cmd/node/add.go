package node

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/urfave/cli/v2"
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

	describe.Nodes(describe.ToNodeChan(node), false)
	return nil
}

func cmdNodeAdd(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateAddNodeOptions(c)
	if err != nil {
		return err
	}

	o := &addNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
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

func readTLSConfigs(c *cli.Context) (caContent, certContent, keyContent string, err error) {
	ca := c.String("ca")
	if ca == "" {
		defaultPath := "/etc/docker/tls/ca.crt"
		if _, err := os.Stat(defaultPath); err == nil {
			ca = defaultPath
		}
	}
	if ca != "" {
		f, err := ioutil.ReadFile(ca)
		if err != nil {
			return "", "", "", fmt.Errorf("Error during reading %s: %v", ca, err)
		}
		caContent = string(f)
	}

	cert := c.String("cert")
	if cert == "" {
		defaultPath := "/etc/docker/tls/client.crt"
		if _, err := os.Stat(defaultPath); err == nil {
			cert = defaultPath
		}
	}
	if cert != "" {
		f, err := ioutil.ReadFile(cert)
		if err != nil {
			return "", "", "", fmt.Errorf("Error during reading %s: %v", cert, err)
		}
		certContent = string(f)
	}

	key := c.String("key")
	if key == "" {
		defaultPath := "/etc/docker/tls/client.key"
		if _, err := os.Stat(defaultPath); err == nil {
			key = defaultPath
		}
	}
	if key != "" {
		f, err := ioutil.ReadFile(key)
		if err != nil {
			return "", "", "", fmt.Errorf("Error during reading %s: %v", key, err)
		}
		keyContent = string(f)
	}
	return caContent, certContent, keyContent, nil
}

func generateAddNodeOptions(c *cli.Context) (*corepb.AddNodeOptions, error) {
	podname := c.Args().First()
	if podname == "" {
		return nil, fmt.Errorf("podname must not be empty")
	}

	nodename := c.String("nodename")

	ca, cert, key, err := readTLSConfigs(c)
	if err != nil {
		return nil, err
	}

	endpoint := c.String("endpoint")
	if endpoint == "" {
		ip := getLocalIP()
		if ip == "" {
			return nil, fmt.Errorf("unable to get local ip")
		}
		port := 2376
		if ca == "" {
			port = 2375
		}
		endpoint = fmt.Sprintf("tcp://%s:%d", ip, port)
	}

	resourceOpts := map[string]*corepb.RawParam{}
	if c.IsSet("cpu") {
		resourceOpts["cpu"] = utils.ToPBRawParamsString(c.String("cpu"))
	}
	if c.IsSet("share") {
		resourceOpts["share"] = utils.ToPBRawParamsString(c.String("share"))
	}
	if c.IsSet("memory") {
		resourceOpts["memory"] = utils.ToPBRawParamsString(c.String("memory"))
	}
	if c.IsSet("storage") {
		resourceOpts["storage"] = utils.ToPBRawParamsString(c.String("storage"))
	}
	if c.IsSet("volumes") {
		resourceOpts["volumes"] = utils.ToPBRawParamsStringSlice(c.StringSlice("volumes"))
	}
	if c.IsSet("numa-cpu") {
		resourceOpts["numa-cpu"] = utils.ToPBRawParamsStringSlice(c.StringSlice("numa-cpu"))
	}
	if c.IsSet("numa-memory") {
		resourceOpts["numa-memory"] = utils.ToPBRawParamsStringSlice(c.StringSlice("numa-memory"))
	}
	if c.IsSet("disk") {
		resourceOpts["disks"] = utils.ToPBRawParamsStringSlice(c.StringSlice("disk"))
	}
	if c.IsSet("workload-limit") {
		resourceOpts["workload-limit"] = utils.ToPBRawParamsStringSlice(c.StringSlice("workload-limit"))
	}

	labels := utils.SplitEquality(c.StringSlice("label"))
	return &corepb.AddNodeOptions{
		Nodename:     nodename,
		Endpoint:     endpoint,
		Podname:      podname,
		Ca:           ca,
		Cert:         cert,
		Key:          key,
		Labels:       labels,
		ResourceOpts: resourceOpts,
	}, nil
}
