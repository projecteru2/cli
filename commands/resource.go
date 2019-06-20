package commands

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

//PodCommand list and add pod
func PodCommand() *cli.Command {
	return &cli.Command{
		Name:  "pod",
		Usage: "pod commands",
		Subcommands: []*cli.Command{
			&cli.Command{
				Name:   "list",
				Usage:  "list all pods",
				Action: listPods,
			},
			&cli.Command{
				Name:      "add",
				Usage:     "add new pod",
				ArgsUsage: podArgsUsage,
				Action:    addPod,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "desc",
						Usage: "description of pod",
						Value: "",
					},
				},
			},
			&cli.Command{
				Name:      "rm",
				Usage:     "remove pod",
				ArgsUsage: podArgsUsage,
				Action:    removePod,
			},
			&cli.Command{
				Name:      "resource",
				Usage:     "pod resource usage",
				ArgsUsage: podArgsUsage,
				Action:    podResource,
			},
			&cli.Command{
				Name:      "nodes",
				Usage:     "list all nodes in one pod",
				ArgsUsage: podArgsUsage,
				Action:    listPodNodes,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "list all nodes or just living nodes",
						Value: false,
					},
				},
			},
			&cli.Command{
				Name:      "networks",
				Usage:     "list all networks in one pod",
				ArgsUsage: podArgsUsage,
				Action:    listPodNetworks,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "driver",
						Usage: "filter driver",
					},
				},
			},
		},
	}
}

func listPods(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()
	resp, err := client.ListPods(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("[ListPods] send request failed %v", err)
	}

	for _, pod := range resp.GetPods() {
		log.Infof("Name: %s, Desc: %s", pod.GetName(), pod.GetDesc())
	}
	return nil
}

func addPod(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()
	desc := c.String("desc")

	pod, err := client.AddPod(context.Background(), &pb.AddPodOptions{
		Name: name,
		Desc: desc,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("[AddPod] success, name: %s, desc: %s", pod.GetName(), pod.GetDesc())
	return nil
}

func removePod(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()

	_, err = client.RemovePod(context.Background(), &pb.RemovePodOptions{
		Name: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("[RemovePod] success, name: %s", name)
	return nil
}

func podResource(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()

	r, err := client.GetPodResource(context.Background(), &pb.GetPodOptions{
		Name: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[PodResource] Pod %s", r.Name)
	for nodename, percent := range r.CpuPercents {
		log.Infof("[PodResource] Node %s Cpu %f%% Memory %f%%", nodename, percent*100, r.MemoryPercents[nodename]*100)
	}
	for nodename, verification := range r.Verifications {
		if verification {
			continue
		}
		log.Warnf("[PodResource] Node %s resource diff %s", nodename, r.Details[nodename])
	}
	return nil
}

func listPodNodes(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()
	all := c.Bool("all")

	resp, err := client.ListPodNodes(context.Background(), &pb.ListNodesOptions{
		Podname: name,
		All:     all,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	for _, node := range resp.GetNodes() {
		log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
	}
	return nil
}

func listPodNetworks(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()
	driver := c.String("driver")

	resp, err := client.ListNetworks(context.Background(), &pb.ListNetworkOptions{
		Podname: name,
		Driver:  driver,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	for _, network := range resp.GetNetworks() {
		log.Infof("Name: %s, Subnets: %s", network.GetName(), strings.Join(network.GetSubnets(), ","))
	}
	return nil
}

//NodeCommand for node control
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
				Name:      "available",
				Usage:     "set availability for a node",
				ArgsUsage: nodeArgsUsage,
				Action:    setNodeAvailable,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "available",
						Usage: "availability",
					},
				},
			},
			&cli.Command{
				Name:      "resource",
				Usage:     "check node resource",
				ArgsUsage: nodeArgsUsage,
				Action:    nodeResource,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "podname",
						Usage: "pod name",
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
					&cli.Int64Flag{
						Name:        "memory",
						Usage:       "memory in bytes",
						DefaultText: "total memory * 0.8",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "add label for node, like a=1 b=2, can set multiple times",
					},
					&cli.StringSliceFlag{
						Name:  "numa-cpu",
						Usage: "numa cpu list, can set multiple times, use comma separated",
					},
					&cli.Int64SliceFlag{
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
		Podname:  "",
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
	for k, v := range node.GetLabels() {
		log.Infof("%s: %s", k, v)
	}
	log.Infof("CPU Used: %f", node.GetCpuUsed())
	log.Infof("Memory Used: %d bytes", node.GetMemoryUsed())
	for nodeID, memory := range node.GetNumaMemory() {
		log.Infof("Memory Node: %s Capacity %d bytes", nodeID, memory)
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
		Podname:  "",
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

func setNodeAvailable(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	name := c.Args().First()
	available := c.Bool("available")

	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Podname:  "",
		Nodename: name,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	_, err = client.SetNodeAvailable(context.Background(), &pb.NodeAvailable{
		Podname:   node.Podname,
		Nodename:  node.Name,
		Available: available,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[SetNodeAvailable] success")
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

	cpu := c.Int("cpu")
	memory := c.Int64("memory")
	numaList := c.StringSlice("numa-cpu")
	numaMemoryList := c.Int64Slice("numa-memory")

	numa := map[string]string{}
	numaMemory := map[string]int64{}

	for index, cpuList := range numaList {
		nodeID := fmt.Sprintf("%d", index)
		for _, cpuID := range strings.Split(cpuList, ",") {
			numa[cpuID] = nodeID
		}
	}

	for index, memory := range numaMemoryList {
		nodeID := fmt.Sprintf("%d", index)
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
	podname := c.String("podname")
	if podname == "" {
		log.Fatal("[NodeResource] no pod name")
	}

	r, err := client.GetNodeResource(context.Background(), &pb.GetNodeOptions{
		Podname: podname, Nodename: nodename,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("[NodeResource] Node %s", r.Name)
	log.Infof("[NodeResource] Cpu %f%% Memory %f%%", r.CpuPercent*100, r.MemoryPercent*100)
	if !r.Verification {
		for _, detail := range r.Details {
			log.Warnf("[PodResource] Resource diff %s", detail)
		}
	}
	return nil
}
