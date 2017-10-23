package commands

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	pb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
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
						Name:  "favor",
						Usage: "name of pod, default: MEM",
						Value: "MEM",
					},
					&cli.StringFlag{
						Name:  "desc",
						Usage: "description of pod",
						Value: "",
					},
				},
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
	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
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
	favor := c.String("favor")
	desc := c.String("desc")

	if favor != coretypes.MEMORY_PRIOR && favor != coretypes.CPU_PRIOR {
		return fmt.Errorf("favor must be MEM/CPU, got %s", favor)
	}

	pod, err := client.AddPod(context.Background(), &pb.AddPodOptions{
		Name:  name,
		Favor: favor,
		Desc:  desc,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	log.Infof("[AddPod] success, name: %s, desc: %s", pod.GetName(), pod.GetDesc())
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
				Name:      "available",
				Usage:     "set availability for a node",
				ArgsUsage: nodeArgsUsage,
				Action:    setNodeAvailable,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "available",
						Usage: "availability",
						Value: true,
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
						Usage: "ca file of docker server",
						Value: "/etc/docker/tls/ca.crt",
					},
					&cli.StringFlag{
						Name:  "cert",
						Usage: "cert file of docker server",
						Value: "/etc/docker/tls/client.crt",
					},
					&cli.StringFlag{
						Name:  "key",
						Usage: "key file of docker server",
						Value: "/etc/docker/tls/client.key",
					},
					&cli.BoolFlag{
						Name:  "public",
						Usage: "set if this node is public",
						Value: false,
					},
					&cli.IntFlag{
						Name:  "cpu",
						Usage: "cpu count",
						Value: 0,
					},
					&cli.Int64Flag{
						Name:  "share",
						Usage: "share count",
						Value: 0,
					},
					&cli.Int64Flag{
						Name:  "memory",
						Usage: "memory in Bytes",
						Value: 0,
					},
				},
			},
		},
	}
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
	for cno, part := range node.GetCpu() {
		log.Infof("Cpu %s has %d capability", cno, part)
	}
	log.Infof("Memory: %d bytes", node.GetMemory())
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

	endpoint := c.String("endpoint")
	if endpoint == "" {
		ip := getLocalIP()
		if ip == "" {
			return cli.Exit(fmt.Errorf("unable to get local ip"), -1)
		}
		endpoint = fmt.Sprintf("tcp://%d:2376", ip)
	}

	ca := c.String("ca")
	caContent, err := ioutil.ReadFile(ca)
	if err != nil {
		return cli.Exit(fmt.Errorf("unable to read %s", ca), -1)
	}

	cert := c.String("cert")
	certContent, err := ioutil.ReadFile(cert)
	if err != nil {
		return cli.Exit(fmt.Errorf("unable to read %s", cert), -1)
	}

	key := c.String("key")
	keyContent, err := ioutil.ReadFile(key)
	if err != nil {
		return cli.Exit(fmt.Errorf("unable to read %s", key), -1)
	}

	share := c.Int64("share")
	if share == 0 {
		share = int64(100)
	}

	cpu := c.Int("cpu")
	memory := c.Int64("memory")

	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	_, err = client.AddNode(context.Background(), &pb.AddNodeOptions{
		Nodename: nodename,
		Endpoint: endpoint,
		Podname:  podname,
		Ca:       string(caContent),
		Cert:     string(certContent),
		Key:      string(keyContent),
		Public:   c.Bool("public"),
		Cpu:      int32(cpu),
		Share:    share,
		Memory:   memory,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}
	log.Infof("[AddNode] success")
	return nil
}
