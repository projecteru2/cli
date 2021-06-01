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

	var f func(*cli.Context, corepb.CoreRPCClient) (*corepb.SetNodeOptions, error)
	if c.Bool("delta") {
		f = generateSetNodeOptionsDelta
	} else {
		f = generateSetNodeOptionsAbsolute
	}

	opts, err := f(c, client)
	if err != nil {
		return err
	}

	o := &setNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateSetNodeOptionsAbsolute(c *cli.Context, client corepb.CoreRPCClient) (*corepb.SetNodeOptions, error) {
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

	var deltaMemory int64 = 0
	if c.IsSet("memory") {
		memory, err := utils.ParseRAMInHuman(c.String("memory"))
		if err != nil {
			return nil, err
		}
		if memory >= 0 {
			deltaMemory = memory - node.InitMemory
		} else {
			return nil, fmt.Errorf("you can't set memory to a negative number when using absolute value")
		}
	}

	var deltaStorage int64 = 0
	if c.IsSet("storage") {
		storage, err := utils.ParseRAMInHuman(c.String("storage"))
		if err != nil {
			return nil, err
		}
		if storage >= 0 {
			deltaStorage = storage - node.InitStorage
		} else {
			return nil, fmt.Errorf("you can't set storage to a negative number when using absolute value")
		}
	}

	return &corepb.SetNodeOptions{
		Nodename:        name,
		StatusOpt:       corepb.TriOpt_KEEP,
		DeltaCpu:        cpuMap,
		DeltaMemory:     deltaMemory,
		DeltaStorage:    deltaStorage,
		DeltaNumaMemory: numaMemory,
		DeltaVolume:     volumeMap,
		Numa:            numa,
		Labels:          utils.SplitEquality(c.StringSlice("label")),
		WorkloadsDown:   c.Bool("mark-workloads-down"),
	}, nil
}

func generateSetNodeOptionsDelta(c *cli.Context, _ corepb.CoreRPCClient) (*corepb.SetNodeOptions, error) {
	name := c.Args().First()
	if name == "" {
		return nil, errors.New("Node name must be given")
	}

	numaMemoryList := c.StringSlice("numa-memory")
	numaMemory := map[string]int64{}
	for nodeID, memoryStr := range numaMemoryList {
		memory, err := utils.ParseRAMInHuman(memoryStr)
		if err != nil {
			return nil, err
		}
		numaMemory[strconv.Itoa(nodeID)] = memory
	}

	numaList := c.StringSlice("numa-cpu")
	numa := map[string]string{}
	for nodeID, cpuList := range numaList {
		for _, cpuID := range strings.Split(cpuList, ",") {
			numa[cpuID] = strconv.Itoa(nodeID)
		}
	}

	cpuList := c.String("cpu")
	cpuMap := map[string]int32{}
	if cpuList != "" {
		cpuMapList := strings.Split(cpuList, ",")
		for _, cpus := range cpuMapList {
			cpuConfigs := strings.Split(cpus, ":")
			// G109: Potential Integer overflow made by strconv.Atoi result conversion to int16/32
			share, err := strconv.Atoi(cpuConfigs[1]) // nolint
			if err != nil {
				return nil, err
			}
			cpuID := cpuConfigs[0]
			cpuMap[cpuID] = int32(share)
		}
	}

	volumeMap := map[string]int64{}
	deltaVolume := c.String("volume")
	if deltaVolume != "" {
		for _, volume := range strings.Split(deltaVolume, ",") {
			parts := strings.Split(volume, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid volume")
			}
			delta, err := utils.ParseRAMInHuman(parts[1])
			if err != nil {
				return nil, err
			}
			volumeMap[parts[0]] = delta
		}
	}

	var (
		deltaMemory  int64
		deltaStorage int64
		err          error
	)
	if deltaMemory, err = utils.ParseRAMInHuman(c.String("memory")); err != nil {
		return nil, err
	}
	if deltaStorage, err = utils.ParseRAMInHuman(c.String("storage")); err != nil {
		return nil, err
	}

	return &corepb.SetNodeOptions{
		Nodename:        name,
		StatusOpt:       corepb.TriOpt_KEEP,
		DeltaCpu:        cpuMap,
		DeltaMemory:     deltaMemory,
		DeltaStorage:    deltaStorage,
		DeltaNumaMemory: numaMemory,
		DeltaVolume:     volumeMap,
		Numa:            numa,
		Labels:          utils.SplitEquality(c.StringSlice("label")),
		WorkloadsDown:   c.Bool("mark-workloads-down"),
	}, nil
}
