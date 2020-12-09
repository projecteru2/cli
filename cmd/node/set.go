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

	opts, err := generateSetNodeOptions(c)
	if err != nil {
		return err
	}

	o := &setNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateSetNodeOptions(c *cli.Context) (*corepb.SetNodeOptions, error) {
	name := c.Args().First()
	if name == "" {
		return nil, errors.New("Node name must be given")
	}

	numaMemoryList := c.StringSlice("delta-numa-memory")
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

	cpuList := c.String("delta-cpu")
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
	deltaVolume := c.String("delta-volume")
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
	if deltaMemory, err = utils.ParseRAMInHuman(c.String("delta-memory")); err != nil {
		return nil, err
	}
	if deltaStorage, err = utils.ParseRAMInHuman(c.String("delta-storage")); err != nil {
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
