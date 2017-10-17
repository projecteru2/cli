package commands

import (
	"fmt"
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
