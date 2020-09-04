package commands

import (
	"encoding/json"
	"fmt"
	"os"

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
		t.AppendHeader(table.Row{"Name", "Cpu", "Memory", "Storage", "Volume"})
		nodeRow := []string{}
		cpuRow := []string{}
		memoryRow := []string{}
		storageRow := []string{}
		volumeRow := []string{}
		for nodename, percent := range r.CpuPercents {
			nodeRow = append(nodeRow, nodename)
			cpuRow = append(cpuRow, fmt.Sprintf("%.2f%%", percent*100))
			memoryRow = append(memoryRow, fmt.Sprintf("%.2f%%", r.MemoryPercents[nodename]*100))
			storageRow = append(storageRow, fmt.Sprintf("%.2f%%", r.StoragePercents[nodename]*100))
			volumeRow = append(volumeRow, fmt.Sprintf("%.2f%%", r.VolumePercents[nodename]*100))
		}
		rows := [][]string{nodeRow, cpuRow, memoryRow, storageRow, volumeRow}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
		t.SetStyle(table.StyleLight)
		t.Render()
	} else {
		log.Infof("[PodResource] Pod %s", r.Name)
		for nodename, percent := range r.CpuPercents {
			log.Infof("[PodResource] Node %s Cpu %.2f%% Memory %.2f%% Storage %.2f%% Volume %.2f%%", nodename, percent*100, r.MemoryPercents[nodename]*100, r.StoragePercents[nodename]*100, r.VolumePercents[nodename]*100)
		}
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
	} else {
		for _, node := range resp.GetNodes() {
			log.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
			r := map[string]interface{}{}
			if err := json.Unmarshal([]byte(node.Info), &r); err != nil {
				log.Errorf("Get Node Info failed: %v", node.Info)
			}
		}
	}
	return nil
}
