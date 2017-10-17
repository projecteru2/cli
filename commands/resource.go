package commands

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

// Pod commands
// list and add pod
func PodCommand() *cli.Command {
	return &cli.Command{
		Name:  "pod",
		Usage: "pod commands",
		SubCommands: []*cli.Command{
			&cli.Command{
				Name:   "list",
				Usage:  "list all pods",
				Action: listPods,
			},
			&cli.Command{
				Name:   "add",
				Usage:  "add new pod",
				Action: addPod,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "name of pod",
					},
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
				Name:   "nodes",
				Usage:  "list all nodes in one pod",
				Action: listPodNodes,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "podname",
						Usage: "name of pod",
					},
					&cli.BoolFlag{
						Name:  "all",
						Usage: "list all nodes or just living nodes",
						Value: false,
					},
				},
			},
			&cli.Command{
				Name:   "networks",
				Usage:  "list all networks in one pod",
				Action: listPodNetworks,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "podname",
						Usage: "name of pod",
					},
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
	name := c.String("name")
	favor := c.String("favor")
	desc := c.String("desc")

	if name == "" {
		log.Fatalf("[AddPod] bad name, got %s", name)
	}
	if favor != "MEM" || favor != "CPU" {
		log.Fatalf("[AddPod] favor must be MEM/CPU, got %s", favor)
	}

	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	pod, err := client.AddPod(context.Background(), &pb.AddPodOptions{
		Name:  name,
		Favor: favor,
		Desc:  desc,
	})
	if err != nil {
		log.Fatalf("[AddPod] send request failed %v", err)
	}

	log.Infof("[AddPod] success, name: %s, desc: %s", pod.GetName(), pod.GetDesc())
	return nil
}

func listPodNodes(c *cli.Context) error {
	name := c.String("podname")
	if name == "" {
		log.Fatalf("[ListPodNodes] bad podname, got %s", name)
	}
	all := c.Bool("all")

	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	resp, err := client.AddPod(context.Background(), &pb.ListNodesOptions{
		Podname: name,
		All:     all,
	})
	if err != nil {
		log.Fatalf("[ListPodNodes] send request failed %v", err)
	}

	for _, node := range resp.GetNodes() {
		log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
	}
	return nil
}

func listPodNetworks(c *cli.Context) error {
	name := c.String("podname")
	if name == "" {
		log.Fatalf("[listPodNetworks] bad podname, got %s", name)
	}
	driver := c.Bool("driver")

	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	resp, err := client.ListNetworks(context.Background(), &pb.ListNetworkOptions{
		Podname: name,
		Driver:  driver,
	})
	if err != nil {
		log.Fatalf("[listPodNetworks] send request failed %v", err)
	}

	for _, network := range resp.GetNetworks() {
		log.Infof("Name: %s, Subnets: %s", node.GetName(), strings.Join(node.GetSubnets(), ","))
	}
	return nil
}

// Node commands
func NodeCommand() *cli.Command {
	return &cli.Command{
		Name:  "node",
		Usage: "node commands",
		SubCommands: []*cli.Command{
			&cli.Command{
				Name:   "get",
				Usage:  "get a node",
				Action: getNode,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "nodename",
						Usage: "name of node",
					},
				},
			},
			&cli.Command{
				Name:   "remove",
				Usage:  "remove a node",
				Action: removeNode,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "nodename",
						Usage: "name of node",
					},
				},
			},
			&cli.Command{
				Name:   "available",
				Usage:  "set availability for a node",
				Action: setNodeAvailable,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "nodename",
						Usage: "name of node",
					},
					&cli.StringFlag{
						Name:  "available",
						Usage: "availability",
						Value: true,
					},
				},
			},
		},
	}
}

func getNode(c *cli.Context) error {
	name := c.String("nodename")
	if name == "" {
		log.Fatalf("[GetNode] bad nodename, got %s", name)
	}

	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Podname:  "",
		Nodename: name,
	})
	if err != nil {
		log.Fatalf("[GetNode] send request failed %v", err)
	}

	log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
	return nil
}

func removeNode(c *cli.Context) error {
	nodename := c.String("nodename")
	if name == "" {
		log.Fatalf("[RemoveNode] bad nodename, got %s", nodename)
	}

	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Podname:  "",
		Nodename: nodename,
	})
	if err != nil {
		log.Fatalf("[RemoveNode] send request failed %v", err)
	}

	_, err := client.RemoveNode(context.Background(), &pb.RemoveNodeOptions{
		Podname:  node.Podname,
		Nodename: node.Nodename,
	})
	if err != nil {
		log.Fatalf("[RemoveNode] send request failed %v", err)
	}
	log.Infof("[RemoveNode] success")
	return nil
}

func setNodeAvailable(c *cli.Context) error {
	name := c.String("nodename")
	if name == "" {
		log.Fatalf("[SetNodeAvailable] bad nodename, got %s", name)
	}
	available := c.Bool("available")

	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	node, err := client.GetNodeByName(context.Background(), &pb.GetNodeOptions{
		Podname:  "",
		Nodename: name,
	})
	if err != nil {
		log.Fatalf("[SetNodeAvailable] send request failed %v", err)
	}

	_, err := client.SetNodeAvailable(context.Background(), &pb.NodeAvailable{
		Podname:   node.Podname,
		Nodename:  node.Name,
		Available: available,
	})
	if err != nil {
		log.Fatalf("[SetNodeAvailable] send request failed %v", err)
	}
	log.Infof("[SetNodeAvailable] success")
	return nil
}
