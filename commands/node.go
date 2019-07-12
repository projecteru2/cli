package commands

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/projecteru2/core/cluster"

	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

// NodeCommand for node control
func NodeCommand() *cli.Command {
	return &cli.Command{
		Name:  "node",
		Usage: "node commands",
		Subcommands: []*cli.Command{
			&cli.Command{
				Name:      "get",
				Usage:     "get a node",
				ArgsUsage: nodeArgsUsage,
				Action:    getNode,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "podname",
						Usage: "which podname",
					},
				},
			},
			&cli.Command{
				Name:      "remove",
				Usage:     "remove a node",
				ArgsUsage: nodeArgsUsage,
				Action:    removeNode,
			},
			&cli.Command{
				Name:      "containers",
				Usage:     "list node containers",
				ArgsUsage: nodeArgsUsage,
				Action:    listNodeContainers,
			},
			&cli.Command{
				Name:      "up",
				Usage:     "set node up",
				ArgsUsage: nodeArgsUsage,
				Action:    setNodeUp,
			},
			&cli.Command{
				Name:      "down",
				Usage:     "set node down",
				ArgsUsage: nodeArgsUsage,
				Action:    setNodeDown,
			},
			&cli.Command{
				Name:      "resource",
				Usage:     "check node resource",
				ArgsUsage: nodeArgsUsage,
				Action:    nodeResource,
			},
			&cli.Command{
				Name:      "set",
				Usage:     "set node resource",
				ArgsUsage: nodeArgsUsage,
				Action:    setNode,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "delta-memory",
						Usage: "memory changes like -1M or 1G, support K, M, G, T",
					},
					&cli.Int64Flag{
						Name:  "delta-storage",
						Usage: "storage changes in bytes",
					},
					&cli.StringFlag{
						Name:  "delta-cpu",
						Usage: "cpu changes in string, like 0:100,1:200,3:50",
					},
					&cli.StringSliceFlag{
						Name:  "delta-numa-memory",
						Usage: "numa memory changes, can set multiple times, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringSliceFlag{
						Name:  "numa-cpu",
						Usage: "numa cpu list, can set multiple times, use comma separated",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "add label for node, like a=1 b=2, can set multiple times",
					},
				},
			},
			&cli.Command{
				Name:      "add",
				Usage:     "add node",
				ArgsUsage: "podname",
				Action:    addNode,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "nodename",
						Usage:   "name of this node, use `hostname` as default",
						EnvVars: []string{"HOSTNAME"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:  "endpoint",
						Usage: "endpoint of docker server",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "ca",
						Usage: "ca file of docker server, like /etc/docker/tls/ca.crt",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "cert",
						Usage: "cert file of docker server, like /etc/docker/tls/client.crt",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "key",
						Usage: "key file of docker server, like /etc/docker/tls/client.key",
						Value: "",
					},
					&cli.IntFlag{
						Name:        "cpu",
						Usage:       "cpu count",
						DefaultText: "total cpu",
					},
					&cli.IntFlag{
						Name:        "share",
						Usage:       "share count",
						DefaultText: "defined in core",
					},
					&cli.StringFlag{
						Name:  "memory",
						Usage: "memory like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "storage",
						Usage: "storage -1M or 1G, support K, M, G, T",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "add label for node, like a=1 b=2, can set multiple times",
					},
					&cli.StringSliceFlag{
						Name:  "numa-cpu",
						Usage: "numa cpu list, can set multiple times, use comma separated",
					},
					&cli.StringSliceFlag{
						Name:  "numa-memory",
						Usage: "numa memory, can set multiple times. if not set, it will count numa-cpu groups, and divided by total memory",
					},
				},
			},
		},
	}
}

func listNodeContainers(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()

	nodename := c.Args().First()
	if nodename == "" {
		return cli.Exit("no node name", -1)
	}
	opts := &pb.GetNodeOptions{
		Nodename: nodename,
	}

	resp, err := client.ListNodeContainers(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for _, container := range resp.Containers {
		log.Infof("%s: %s", container.Name, container.Id)
		log.Infof("Pod %s, Node %s, CPU %v, Quota %v, Memory %v, Privileged %v", container.Podname, container.Nodename, container.Cpu, container.Quota, container.Memory, container.Privileged)
	}
	return nil
}

func getNode(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()

	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
	for k, v := range node.GetLabels() {
		log.Infof("%s: %s", k, v)
	}
	log.Infof("CPU Used: %.2f", node.GetCpuUsed())
	log.Infof("Memory Used: %d bytes", node.GetMemoryUsed())
	for nodeID, memory := range node.GetNumaMemory() {
		log.Infof("Memory Node: %s Capacity %d bytes", nodeID, memory)
	}

	if node.GetInitStorage() > 0 {
		log.Infof("Storage Used: %d bytes", node.GetStorageUsed())
	} else {
		log.Infof("Storage Used: %d bytes (unlimited)", node.GetStorageUsed())
	}
	return nil
}

func removeNode(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()
	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	_, err = client.RemoveNode(context.Background(), &pb.RemoveNodeOptions{
		Podname:  node.Podname,
		Nodename: node.Name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[RemoveNode] success")
	return nil
}

func setNode(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()

	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	numaMemoryList := c.StringSlice("delta-numa-memory")
	numaMemory := map[string]int64{}

	for index, memoryStr := range numaMemoryList {
		var memory int64
		nodeID := fmt.Sprintf("%d", index)
		if memory, err = parseRAMInHuman(memoryStr); err != nil {
			return cli.Exit(err, -1)
		}
		numaMemory[nodeID] = memory
	}

	numaList := c.StringSlice("numa-cpu")
	numa := map[string]string{}

	for index, cpuList := range numaList {
		nodeID := fmt.Sprintf("%d", index)
		for _, cpuID := range strings.Split(cpuList, ",") {
			numa[cpuID] = nodeID
		}
	}

	labels := map[string]string{}
	for _, d := range c.StringSlice("label") {
		parts := strings.SplitN(d, "=", 2)
		if len(parts) != 2 {
			continue
		}
		labels[parts[0]] = parts[1]
	}

	cpuList := c.String("delta-cpu")
	cpuMap := map[string]int32{}
	if cpuList != "" {
		cpuMapList := strings.Split(cpuList, ",")
		for _, cpus := range cpuMapList {
			cpuConfigs := strings.Split(cpus, ":")
			share, err := strconv.Atoi(cpuConfigs[1])
			if err != nil {
				return cli.Exit(err, -1)
			}
			cpuID := cpuConfigs[0]
			cpuMap[cpuID] = int32(share)
		}
	}

	var deltaMemory int64
	if deltaMemory, err = parseRAMInHuman(c.String("delta-memory")); err != nil {
		return cli.Exit(err, -1)
	}

	_, err = client.SetNode(context.Background(), &pb.SetNodeOptions{
		Podname:         node.Podname,
		Nodename:        node.Name,
		Status:          cluster.KeepNodeStatus,
		DeltaCpu:        cpuMap,
		DeltaMemory:     deltaMemory,
		DeltaStorage:    c.Int64("delta-storage"),
		DeltaNumaMemory: numaMemory,
		Numa:            numa,
		Labels:          labels,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[SetNode] set node %s success", name)
	return nil
}

func setNodeUp(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()

	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	_, err = client.SetNode(context.Background(), &pb.SetNodeOptions{
		Podname:  node.Podname,
		Nodename: node.Name,
		Status:   cluster.NodeUp,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[SetNode] node %s up", name)
	return nil
}

func setNodeDown(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()

	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	_, err = client.SetNode(context.Background(), &pb.SetNodeOptions{
		Podname:  node.Podname,
		Nodename: node.Name,
		Status:   cluster.NodeDown,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[SetNode] node %s down", name)
	return nil
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

func addNode(c *cli.Context) error {
	podname := c.Args().First()
	if podname == "" {
		return cli.Exit(fmt.Errorf("podname must not be empty"), -1)
	}

	nodename := c.String("nodename")
	if nodename == "" {
		n, err := os.Hostname()
		if err != nil {
			return cli.Exit(err, -1)
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
			return cli.Exit(fmt.Errorf("Error during reading %s: %v", ca, err), -1)
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
			return cli.Exit(fmt.Errorf("Error during reading %s: %v", cert, err), -1)
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
			return cli.Exit(fmt.Errorf("Error during reading %s: %v", key, err), -1)
		}
		keyContent = string(f)
	}

	endpoint := c.String("endpoint")
	if endpoint == "" {
		ip := getLocalIP()
		if ip == "" {
			return cli.Exit(fmt.Errorf("unable to get local ip"), -1)
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

	var err error
	var memory, storage int64
	if memory, err = parseRAMInHuman(c.String("memory")); err != nil {
		return cli.Exit(err, -1)
	}
	if storage, err = parseRAMInHuman(c.String("storage")); err != nil {
		return cli.Exit(err, -1)
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
		memory, err := parseRAMInHuman(memoryStr)
		if err != nil {
			return cli.Exit(err, -1)
		}
		numaMemory[nodeID] = memory
	}

	labels := map[string]string{}
	for _, d := range c.StringSlice("label") {
		parts := strings.SplitN(d, "=", 2)
		if len(parts) != 2 {
			continue
		}
		labels[parts[0]] = parts[1]
	}

	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	resp, err := client.AddNode(context.Background(), &pb.AddNodeOptions{
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
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[AddNode] success")
	log.Infof("%s add %s at %s", podname, nodename, resp.Endpoint)
	for k, v := range resp.Labels {
		log.Infof("%s: %s", k, v)
	}
	return nil
}

func nodeResource(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	nodename := c.Args().First()

	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Nodename: nodename,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	r, err := client.GetNodeResource(context.Background(), &pb.GetNodeOptions{
		Podname: node.Podname, Nodename: node.Name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("[NodeResource] Node %s", r.Name)
	log.Infof("[NodeResource] Cpu %.2f%% Memory %.2f%% Storage %.2f%%", r.CpuPercent*100, r.MemoryPercent*100, r.StoragePercent*100)
	if !r.Verification {
		for _, detail := range r.Details {
			log.Warnf("[PodResource] Resource diff %s", detail)
		}
	}
	return nil
}
