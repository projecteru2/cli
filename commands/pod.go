package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

// PodCommand list and add pod
func PodCommand() *cli.Command {
	return &cli.Command{
		Name:  "pod",
		Usage: "pod commands",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list all pods",
				Action: listPods,
			},
			{
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
			{
				Name:      "remove",
				Usage:     "remove pod",
				ArgsUsage: podArgsUsage,
				Action:    removePod,
			},
			{
				Name:      "resource",
				Usage:     "pod resource usage",
				ArgsUsage: podArgsUsage,
				Action:    podResource,
			},
			{
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
			{
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
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	resp, err := client.ListPods(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("[ListPods] send request failed %v", err)
	}

	if c.Bool("pretty") {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Description"})
		nameRow := []string{}
		descRow := []string{}
		for _, pod := range resp.GetPods() {
			nameRow = append(nameRow, pod.Name)
			descRow = append(descRow, pod.Desc)

		}
		rows := [][]string{nameRow, descRow}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
		t.SetStyle(table.StyleLight)
		t.Render()
	} else {
		for _, pod := range resp.GetPods() {
			log.Infof("Name: %s, Desc: %s", pod.GetName(), pod.GetDesc())
		}
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
	if c.Bool("pretty") {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Cpu", "Memory", "Storage", "Volume", "Diffs"})
		nodeRow := []string{}
		cpuRow := []string{}
		memoryRow := []string{}
		storageRow := []string{}
		volumeRow := []string{}
		diffsRow := []string{}
		for _, nodeResource := range r.NodesResource {
			nodeRow = append(nodeRow, nodeResource.Name)
			cpuRow = append(cpuRow, fmt.Sprintf("%.2f%%", nodeResource.CpuPercent*100))
			memoryRow = append(memoryRow, fmt.Sprintf("%.2f%%", nodeResource.MemoryPercent*100))
			storageRow = append(storageRow, fmt.Sprintf("%.2f%%", nodeResource.StoragePercent*100))
			volumeRow = append(volumeRow, fmt.Sprintf("%.2f%%", nodeResource.VolumePercent*100))
			diffsRow = append(diffsRow, strings.Join(nodeResource.Diffs, "\n"))
		}
		rows := [][]string{nodeRow, cpuRow, memoryRow, storageRow, volumeRow, diffsRow}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
		t.SetStyle(table.StyleLight)
		t.Render()
	} else {
		log.Infof("[PodResource] Pod %s", r.Name)
		for _, nodeResource := range r.NodesResource {
			log.Infof("[PodResource] Node %s Cpu %.2f%% Memory %.2f%% Storage %.2f%% Volume %.2f%%",
				nodeResource.Name, nodeResource.CpuPercent*100, nodeResource.MemoryPercent*100,
				nodeResource.StoragePercent*100, nodeResource.VolumePercent*100,
			)
			if len(nodeResource.Diffs) > 0 {
				log.Warnf("[PodResource] Node %s resource diff %s", nodeResource.Name, strings.Join(nodeResource.Diffs, "\n"))
			}
		}
	}
	return nil
}

func listPodNodes(c *cli.Context) error {
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	name := c.Args().First()
	all := c.Bool("all")

	resp, err := client.ListPodNodes(context.Background(), &pb.ListNodesOptions{
		Podname: name,
		All:     all,
	})
	if err != nil {
		return cli.Exit(err, -1)
	}

	if c.Bool("pretty") {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Endpoint"})
		nameRow := []string{}
		endpointRow := []string{}
		for _, node := range resp.Nodes {
			nameRow = append(nameRow, node.Name)
			endpointRow = append(endpointRow, node.Endpoint)
		}
		rows := [][]string{nameRow, endpointRow}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
		t.SetStyle(table.StyleLight)
		t.Render()
		return nil
	}

	for _, node := range resp.GetNodes() {
		log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
		r := map[string]interface{}{}
		if err := json.Unmarshal([]byte(node.Info), &r); err != nil {
			log.Errorf("Get Node Info failed: %v", node.Info)
		}
	}
	return nil
}
