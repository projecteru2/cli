package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type setNodeOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.SetNodeOptions
}

func (o *setNodeOptions) run(ctx context.Context) error {
	_, err := o.client.SetNode(ctx, o.opts)
	if err != nil {
		return err
	}
	logrus.Infof("[SetNode] set node %s success", o.opts.Nodename)
	return nil
}

func cmdNodeSet(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateSetNodeOptions(c, client)
	if err != nil {
		return err
	}

	o := &setNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateSetNodeOptions(c *cli.Context, client corepb.CoreRPCClient) (*corepb.SetNodeOptions, error) {
	name := c.Args().First()
	if name == "" {
		return nil, errors.New("Node name must be given")
	}

	node, err := client.GetNode(c.Context, &corepb.GetNodeOptions{Nodename: name})
	if err != nil {
		return nil, err
	}

	numaMemoryList := c.StringSlice("numa-memory")
	numaMemory := map[string]int64{}
	for index, memoryStr := range numaMemoryList {
		memory, err := utils.ParseRAMInHuman(memoryStr)
		if err != nil {
			return nil, err
		}
		nodeID := strconv.Itoa(index)
		numaMemory[nodeID] = memory - node.InitNumaMemory[nodeID]
	}

	numaList := c.StringSlice("numa-cpu")
	numa := map[string]string{}
	for index, cpuList := range numaList {
		nodeID := strconv.Itoa(index)
		for _, cpuID := range strings.Split(cpuList, ",") {
			numa[cpuID] = nodeID
		}
	}

	cpuMap := map[string]int32{}
	cpuList := c.String("cpu")
	if cpuList != "" {
		for _, cpus := range strings.Split(cpuList, ",") {
			cpuConfigs := strings.Split(cpus, ":")
			if len(cpuConfigs) != 2 {
				return nil, fmt.Errorf("invalid cpu share")
			}
			// G109: Potential Integer overflow made by strconv.Atoi result conversion to int16/32
			share, err := strconv.Atoi(cpuConfigs[1]) // nolint
			if err != nil {
				return nil, err
			}
			cpuID := cpuConfigs[0]
			cpuMap[cpuID] = int32(share) - node.InitCpu[cpuID]
		}
	}

	volumeMap := map[string]int64{}
	volumes := c.String("volume")
	if volumes != "" {
		for _, volume := range strings.Split(volumes, ",") {
			parts := strings.Split(volume, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid volume")
			}

			v, err := utils.ParseRAMInHuman(parts[1])
			if err != nil {
				return nil, err
			}

			name := parts[0]
			volumeMap[name] = v - node.InitVolume[name]
		}
	}

	memory, err := utils.ParseRAMInHuman(c.String("memory"))
	if err != nil {
		return nil, err
	}
	if memory > 0 {
		memory = memory - node.InitMemory
	}

	storage, err := utils.ParseRAMInHuman(c.String("storage"))
	if err != nil {
		return nil, err
	}
	if storage > 0 {
		storage = storage - node.InitStorage
	}

	return &corepb.SetNodeOptions{
		Nodename:        name,
		StatusOpt:       corepb.TriOpt_KEEP,
		DeltaCpu:        cpuMap,
		DeltaMemory:     memory,
		DeltaStorage:    storage,
		DeltaNumaMemory: numaMemory,
		DeltaVolume:     volumeMap,
		Numa:            numa,
		Labels:          utils.SplitEquality(c.StringSlice("label")),
		WorkloadsDown:   c.Bool("mark-workloads-down"),
	}, nil
}
