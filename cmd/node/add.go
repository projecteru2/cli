package node

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

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

	describe.Nodes(node)
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

func generateAddNodeOptions(c *cli.Context) (*corepb.AddNodeOptions, error) {
	podname := c.Args().First()
	if podname == "" {
		return nil, fmt.Errorf("podname must not be empty")
	}

	nodename := c.String("nodename")
	if nodename == "" {
		n, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		nodename = n
	}

	ca := c.String("ca")
	if ca == "" {
		defaultPath := "/etc/docker/tls/ca.crt"
		if _, err := os.Stat(defaultPath); err == nil {
			ca = defaultPath
		}
	}
	caContent := ""
	if ca != "" {
		f, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("Error during reading %s: %v", ca, err)
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
	certContent := ""
	if cert != "" {
		f, err := ioutil.ReadFile(cert)
		if err != nil {
			return nil, fmt.Errorf("Error during reading %s: %v", cert, err)
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
	keyContent := ""
	if key != "" {
		f, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, fmt.Errorf("Error during reading %s: %v", key, err)
		}
		keyContent = string(f)
	}

	endpoint := c.String("endpoint")
	if endpoint == "" {
		ip := getLocalIP()
		if ip == "" {
			return nil, fmt.Errorf("unable to get local ip")
		}
		port := 2376
		if caContent == "" {
			port = 2375
		}
		endpoint = fmt.Sprintf("tcp://%s:%d", ip, port)
	}

	share := c.Int("share")
	if share == 0 {
		share = 100
	}

	var (
		err             error
		memory, storage int64
	)
	if memory, err = utils.ParseRAMInHuman(c.String("memory")); err != nil {
		return nil, err
	}
	if storage, err = utils.ParseRAMInHuman(c.String("storage")); err != nil {
		return nil, err
	}

	cpu := c.Int("cpu")

	numa := map[string]string{}
	numaMemory := map[string]int64{}

	for index, cpuList := range c.StringSlice("numa-cpu") {
		nodeID := fmt.Sprintf("%d", index)
		for _, cpuID := range strings.Split(cpuList, ",") {
			numa[cpuID] = nodeID
		}
	}

	for index, memoryStr := range c.StringSlice("numa-memory") {
		nodeID := fmt.Sprintf("%d", index)
		memory, err := utils.ParseRAMInHuman(memoryStr)
		if err != nil {
			return nil, err
		}
		numaMemory[nodeID] = memory
	}

	volumes := map[string]int64{}

	for _, volume := range c.StringSlice("volumes") {
		parts := strings.Split(volume, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid volume")
		}

		capacity, err := utils.ParseRAMInHuman(parts[1])
		if err != nil {
			return nil, err
		}
		volumes[parts[0]] = capacity
	}

	labels := utils.SplitEquality(c.StringSlice("label"))
	return &corepb.AddNodeOptions{
		Nodename:   nodename,
		Endpoint:   endpoint,
		Podname:    podname,
		Ca:         caContent,
		Cert:       certContent,
		Key:        keyContent,
		Cpu:        int32(cpu),
		Share:      int32(share),
		Memory:     memory,
		Storage:    storage,
		Labels:     labels,
		Numa:       numa,
		NumaMemory: numaMemory,
		VolumeMap:  volumes,
	}, nil
}
