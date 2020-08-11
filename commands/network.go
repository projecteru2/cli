package commands

import (
	"context"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// NetworkCommand list and add pod
func NetworkCommand() *cli.Command {
	return &cli.Command{
		Name:  "network",
		Usage: "network commands",
		Subcommands: []*cli.Command{
			{
				Name:      "list",
				ArgsUsage: podArgsUsage,
				Usage:     "list one pod all networks",
				Action:    listPodNetworks,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "driver",
						Usage: "filter driver",
					},
				},
			},
			{
				Name:      "connect",
				ArgsUsage: containerArgsUsage,
				Usage:     "connect containers to network",
				Action:    connectToNetwork,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "network",
						Usage:    "network name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "ipv4",
						Usage: "specify ipv4",
					},
					&cli.StringFlag{
						Name:  "ipv6",
						Usage: "specify ipv6",
					},
				},
			},
			{
				Name:      "disconnect",
				ArgsUsage: containerArgsUsage,
				Usage:     "disconnect containers to network",
				Action:    disconnectToNetwork,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "network",
						Usage:    "network name",
						Required: true,
					},
				},
			},
		},
	}
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

	if c.Bool("pretty") {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Network"})
		nameRow := []string{}
		networkRow := []string{}
		for _, network := range resp.Networks {
			nameRow = append(nameRow, network.Name)
			networkRow = append(networkRow, strings.Join(network.GetSubnets(), ","))
		}
		rows := [][]string{nameRow, networkRow}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
		t.SetStyle(table.StyleLight)
		t.Render()
	} else {
		for _, network := range resp.GetNetworks() {
			log.Infof("Name: %s, Subnets: %s", network.GetName(), strings.Join(network.GetSubnets(), ","))
		}
	}
	return nil
}

func connectToNetwork(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	network := c.String("network")
	ipv4 := c.String("ipv4")
	ipv6 := c.String("ipv6")

	IDs := c.Args().Slice()
	for _, ID := range IDs {
		resp, err := client.ConnectNetwork(c.Context, &pb.ConnectNetworkOptions{
			Network: network,
			Target:  ID,
			Ipv4:    ipv4,
			Ipv6:    ipv6,
		})
		if err != nil {
			log.Warnf("[connectToNetwork] Connect %s to network %s failed", ID, network)
		} else {
			log.Infof("[connectToNetwork] Connect %s at %v", ID, resp.Subnets)
		}
	}
	return nil
}

func disconnectToNetwork(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	network := c.String("network")

	IDs := c.Args().Slice()
	for _, ID := range IDs {
		if _, err := client.DisconnectNetwork(c.Context, &pb.DisconnectNetworkOptions{
			Network: network,
			Target:  ID,
		}); err != nil {
			log.Warnf("[disConnectToNetwork] Disconnect %s to network %s failed", ID, network)
		} else {
			log.Infof("[disConnectToNetwork] Disconnect %s success", ID)
		}
	}
	return nil
}
