package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/jedib0t/go-pretty/table"
	"github.com/projecteru2/cli/utils"
	"github.com/projecteru2/core/cluster"
	pb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/strategy"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

const unlimited = "UNLIMITED"

// ContainerCommand for control containers
func ContainerCommand() *cli.Command {
	return &cli.Command{
		Name:  "container",
		Usage: "container commands",
		Subcommands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "get container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    getContainers,
			},
			{
				Name:      "logs",
				Usage:     "get container stream logs",
				ArgsUsage: "containerID",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "tail",
						Value: "all",
						Usage: "how many",
					},
				},
				Action: getContainerLog,
			},
			{
				Name:      "get-status",
				Usage:     "get container status",
				ArgsUsage: containerArgsUsage,
				Action:    getContainersStatus,
			},
			{
				Name:      "set-status",
				Usage:     "set container status",
				ArgsUsage: containerArgsUsage,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "running",
						Usage: "Running",
					},
					&cli.BoolFlag{
						Name:  "healthy",
						Usage: "Healthy",
					},
					&cli.Int64Flag{
						Name:  "ttl",
						Usage: "ttl",
						Value: 0,
					},
					&cli.StringSliceFlag{
						Name:  "network",
						Usage: "network, can set multiple times, name=ip",
					},
					&cli.StringFlag{
						Name:  "extension",
						Usage: "extension things",
					},
				},
				Action: setContainersStatus,
			},
			{
				Name:      "list",
				Usage:     "list container(s) by appname",
				ArgsUsage: "[appname]",
				Action:    listContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "entry",
						Usage: "filter by entry",
					},
					&cli.StringFlag{
						Name:  "nodename",
						Usage: "filter by nodename",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "label filter can set multiple times",
					},
					&cli.Int64Flag{
						Name:  "limit",
						Usage: "limit data size",
					},
				},
			},
			{
				Name:      "stop",
				Usage:     "stop container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    stopContainers,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to stop",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			{
				Name:      "start",
				Usage:     "start container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    startContainers,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to start",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			{
				Name:      "restart",
				Usage:     "restart container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    restartContainers,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to restart",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "remove container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    removeContainers,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to remove",
						Aliases: []string{"f"},
						Value:   false,
					},
					&cli.IntFlag{
						Name:    "step",
						Usage:   "concurrent remove step",
						Aliases: []string{"s"},
						Value:   1,
					},
				},
			},
			{
				Name:      "copy",
				Usage:     "copy file(s) from container(s)",
				ArgsUsage: copyArgsUsage,
				Action:    copyContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Usage:   "where to store",
						Aliases: []string{"d"},
						Value:   "/tmp",
					},
				},
			},
			{
				Name:      "send",
				Usage:     "send file(s) to container(s)",
				ArgsUsage: sendArgsUsage,
				Action:    sendContainers,
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "file",
						Usage: "copy local files to container, can use multiple times. src_path:dst_path",
					},
				},
			},
			{
				Name:      "dissociate",
				Usage:     "Dissociate container(s) from eru, return it resource but not remove it",
				ArgsUsage: containerArgsUsage,
				Action:    dissociateContainers,
			},
			{
				Name:      "realloc",
				Usage:     "realloc containers resource",
				ArgsUsage: containerArgsUsage,
				Action:    reallocContainers,
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:  "cpu-request",
						Usage: "cpu request increment/decrement",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  "cpu-limit",
						Usage: "cpu limit increment/decrement",
						Value: 0,
					},
					&cli.StringFlag{
						Name:  "memory-request",
						Usage: "memory request increment/decrement, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "memory-limit",
						Usage: "memory limit increment/decrement, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "volumes-request",
						Usage: `volumes request increment/decrement, like "AUTO:/data:rw:-1G,/tmp:/tmp"`,
					},
					&cli.StringFlag{
						Name:  "volumes-limit",
						Usage: `volumes limit increment/decrement, like "AUTO:/data:rw:-1G,/tmp:/tmp"`,
					},
					&cli.BoolFlag{
						Name:  "cpu-bind",
						Usage: `bind fixed cpu(s) with container`,
					},
					&cli.BoolFlag{
						Name:  "cpu-unbind",
						Usage: `unbind the container relation with cpu`,
					},
					&cli.StringFlag{
						Name:  "storage-request",
						Usage: `storage request incr/decr, like "-1G"`,
					},
					&cli.StringFlag{
						Name:  "storage-limit",
						Usage: `storage limit incr/decr, like "-1G"`,
					},
				},
			},
			{
				Name:      "exec",
				Usage:     "run a command in a running container",
				ArgsUsage: "containerID -- cmd1 cmd2 cmd3",
				Action:    execContainer,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"i"},
						Value:   false,
					},
					&cli.StringSliceFlag{
						Name:    "env",
						Aliases: []string{"e"},
						Usage:   "ENV=value",
					},
					&cli.StringFlag{
						Name:    "workdir",
						Aliases: []string{"w"},
						Usage:   "/path/to/workdir",
						Value:   "/",
					},
				},
			},
			{
				Name:      "replace",
				Usage:     "replace containers by params",
				ArgsUsage: specFileURI,
				Action:    replaceContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pod",
						Usage: "where to replace",
					},
					&cli.StringFlag{
						Name:  "entry",
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  "image",
						Usage: "which to replace",
					},
					&cli.StringFlag{
						Name:  "node",
						Usage: "which node to replace",
						Value: "",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "run simultaneously",
						Value: 1,
					},
					&cli.BoolFlag{
						Name:  "network-inherit",
						Usage: "use old container network configuration",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "SDN name or host mode",
						//	Value: "host",
					},
					&cli.StringSliceFlag{
						Name:  "env",
						Usage: "set env can use multiple times, e.g., GO111MODULE=on",
					},
					&cli.StringFlag{
						Name:  "user",
						Usage: "which user",
						Value: "root",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "filter container by labels",
					},
					&cli.StringSliceFlag{
						Name:  "file",
						Usage: "copy local files to container, can use multiple times. src_path:dst_path",
					},
					&cli.StringSliceFlag{
						Name:  "copy",
						Usage: "copy old container files to new container, can use multiple times. src_path:dst_path",
					},
					&cli.BoolFlag{
						Name:  "debug",
						Usage: "enable debug mode for container send their logs to default log driver",
					},
					&cli.BoolFlag{
						Name:  "ignore-hook",
						Usage: "ignore-hook result",
						Value: false,
					},
					&cli.StringSliceFlag{
						Name:  "after-create",
						Usage: "run commands after create",
					},
				},
			},
			{
				Name:      "deploy",
				Usage:     "deploy containers by params",
				ArgsUsage: specFileURI,
				Action:    deployContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pod",
						Usage: "where to run",
					},
					&cli.StringFlag{
						Name:  "entry",
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  "image",
						Usage: "which to run",
					},
					&cli.StringSliceFlag{
						Name:  "node",
						Usage: "which node to run",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "how many",
						Value: 1,
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "SDN name or host mode",
						Value: "host",
					},
					&cli.Float64Flag{
						Name:  "cpu-request",
						Usage: "how many cpu to request",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  "cpu-limit",
						Usage: "how many cpu to limit; can specify limit without request",
						Value: 1.0,
					},
					&cli.StringFlag{
						Name:  "memory-request",
						Usage: "how many memory to request like 1M or 1G, support K, M, G, T",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "memory-limit",
						Usage: "how many memory to limit like 1M or 1G, support K, M, G, T; can specify limit without request",
						Value: "512M",
					},
					&cli.StringFlag{
						Name:  "storage-request",
						Usage: "how many storage to request quota like 1M or 1G, support K, M, G, T",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "storage-limit",
						Usage: "how many storage to limit quota like 1M or 1G, support K, M, G, T; can specify limit without request",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  "env",
						Usage: "set env can use multiple times, e.g., GO111MODULE=on",
					},
					&cli.StringSliceFlag{
						Name:  "nodelabel",
						Usage: "filter nodes by labels",
					},
					&cli.StringFlag{
						Name:  "deploy-strategy",
						Usage: "deploy method auto/fill/each/global",
						Value: strategy.Auto,
					},
					&cli.StringFlag{
						Name:  "user",
						Usage: "which user",
						Value: "root",
					},
					&cli.StringSliceFlag{
						Name:  "file",
						Usage: "copy local file to container, can use multiple times. src_path:dst_path",
					},
					&cli.StringSliceFlag{
						Name:  "after-create",
						Usage: "run commands after create",
					},
					&cli.BoolFlag{
						Name:  "debug",
						Usage: "enable debug mode for container send their logs to default log driver",
					},
					&cli.IntFlag{
						Name:  "nodes-limit",
						Usage: "Limit nodes count in fill and each mode",
						Value: 0,
					},
					&cli.BoolFlag{
						Name:  "auto-replace",
						Usage: "create or replace automatically",
					},
					&cli.BoolFlag{
						Name:  "cpu-bind",
						Usage: "bind cpu or not",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "ignore-hook",
						Usage: "ignore hook process",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "raw-args",
						Usage: "raw args in json (for docker engine)",
						Value: "",
					},
				},
			},
		},
	}
}

func removeContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	opts := &pb.RemoveContainerOptions{Ids: c.Args().Slice(), Force: c.Bool("force"), Step: int32(c.Int("step"))}
	if opts.Force {
		log.Warn("If container not stopped, force to remove will not trigger hook process if set")
	}
	resp, err := client.RemoveContainer(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Success {
			log.Infof("[RemoveContainer] %s Success", msg.Id)
		} else {
			log.Errorf("[RemoveContainer] %s Failed", msg.Id)
		}
		if msg.Hook != "" {
			log.Info(msg.Hook)
		}
	}
	return nil
}

func renderContainer(container *pb.Container) {
	cpuRequest, cpuLimit := unlimited, unlimited
	if container.Resource.CpuQuotaRequest != 0 {
		cpuRequest = fmt.Sprintf("%v", container.Resource.CpuQuotaRequest)
	}
	if container.Resource.CpuQuotaLimit != 0 {
		cpuLimit = fmt.Sprintf("%v", container.Resource.CpuQuotaLimit)
	}
	memoryRequest, memoryLimit := unlimited, unlimited
	if container.Resource.MemoryRequest != 0 {
		memoryRequest = units.HumanSize(float64(container.Resource.MemoryRequest))
	}
	if container.Resource.MemoryLimit != 0 {
		memoryLimit = units.HumanSize(float64(container.Resource.MemoryLimit))
	}
	storageRequest, storageLimit := unlimited, unlimited
	if container.Resource.StorageRequest != 0 {
		storageRequest = units.HumanSize(float64(container.Resource.StorageRequest))
	}
	if container.Resource.StorageLimit != 0 {
		storageLimit = units.HumanSize(float64(container.Resource.StorageLimit))
	}
	log.Info("--------------------------------------")
	log.Infof("%s: %s", container.Name, container.Id)
	log.Infof("Pod: %s, Node: %s", container.Podname, container.Nodename)
	log.Infof("QuotaRequest: %v, QuotaLimit: %v, CPUMap: %v", cpuRequest, cpuLimit, container.Resource.Cpu)
	log.Infof("MemoryRequest: %v, MemoryLimit: %v", memoryRequest, memoryLimit)
	log.Infof("StorageRequest: %v, StorageLimit: %v", storageRequest, storageLimit)
	log.Infof("VolumeRequest: %v, VolumeLimit: %v", container.Resource.VolumesRequest, container.Resource.VolumesLimit)
	log.Infof("VolumePlanRequest: %+v, VolumePlanLimit: %+v, Privileged %v", container.Resource.VolumePlanRequest, container.Resource.VolumePlanLimit, container.Privileged)
	for networkName, IP := range container.Publish {
		log.Infof("Publish at %s ip %s", networkName, IP)
	}
	if container.Status == nil {
		log.Warn("Can't get container status, maybe dissociate with Eru")
	} else {
		log.Infof("Networks: %v", container.Status.Networks)
	}
}

func prettyRenderContianers(containers []*pb.Container) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name/ID", "Pod", "Node", "Status", "Volume", "IP", "Networks"})

	for _, c := range containers {
		// publish ip
		ips := []string{}
		for networkName, IP := range c.Publish {
			ips = append(ips, fmt.Sprintf("%s: %s", networkName, IP))
		}

		// networks
		ns := []string{}
		if c.Status != nil {
			for name, ip := range c.Status.Networks {
				ns = append(ns, fmt.Sprintf("%s: %s", name, ip))
			}
		}

		rows := [][]string{
			{c.Name, c.Id},
			{c.Podname},
			{c.Nodename},
			{fmt.Sprintf("QuotaRequest: %f", c.Resource.CpuQuotaRequest), fmt.Sprintf("QuotaLimit: %f", c.Resource.CpuQuotaLimit), fmt.Sprintf("MemoryRequest: %v", c.Resource.MemoryRequest), fmt.Sprintf("MemoryLimit: %v", c.Resource.MemoryLimit), fmt.Sprintf("StorageRequest: %v", c.Resource.StorageRequest), fmt.Sprintf("StorageLimit: %v", c.Resource.StorageLimit), fmt.Sprintf("Privileged: %v", c.Privileged)},
			c.Resource.VolumesRequest,
			c.Resource.VolumesLimit,
			{fmt.Sprintf("VolumePlanRequest: %+v", c.Resource.VolumePlanRequest)},
			{fmt.Sprintf("VolumePlanLimit: %+v", c.Resource.VolumePlanLimit)},
			ips,
			ns,
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func renderContainerStatus(containerStatus *pb.ContainerStatus) {
	log.Info("--------------------------------------")
	log.Infof("ID: %s", containerStatus.Id)
	log.Infof("Running: %v, Healthy: %v", containerStatus.Running, containerStatus.Healthy)
	log.Infof("Networks: %v", containerStatus.Networks)
	log.Infof("Extension %s", containerStatus.Extension)
}

func prettyRenderContainerStatus(containerStatuses []*pb.ContainerStatus) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Status", "Networks", "Extensions"})

	for _, s := range containerStatuses {
		// networks
		ns := []string{}
		for name, ip := range s.Networks {
			ns = append(ns, fmt.Sprintf("%s: %s", name, ip))
		}

		// extensions
		extensions := map[string]string{}
		if err := json.Unmarshal(s.Extension, &extensions); err != nil {
			log.Errorf("json unmarshal failed %v", err)
			continue
		}
		es := []string{}
		for k, v := range extensions {
			es = append(es, fmt.Sprintf("%s: %s", k, v))
		}

		rows := [][]string{
			{s.Id},
			{fmt.Sprintf("Running: %v", s.Running), fmt.Sprintf("Healthy: %v", s.Healthy)},
			ns,
			es,
		}
		t.AppendRows(toTableRows(rows))
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func getContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	resp, err := client.GetContainers(context.Background(), &pb.ContainerIDs{Ids: c.Args().Slice()})
	if err != nil {
		return cli.Exit(err, -1)
	}

	if c.Bool("pretty") {
		prettyRenderContianers(resp.Containers)
	} else {
		for _, container := range resp.Containers {
			renderContainer(container)
		}
	}
	return nil
}

func getContainersStatus(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	resp, err := client.GetContainersStatus(context.Background(), &pb.ContainerIDs{Ids: c.Args().Slice()})
	if err != nil {
		return cli.Exit(err, -1)
	}

	if c.Bool("pretty") {
		prettyRenderContainerStatus(resp.Status)
	} else {
		for _, containerStatus := range resp.Status {
			renderContainerStatus(containerStatus)
		}
	}
	return nil
}

func setContainersStatus(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	running := c.Bool("running")
	healthy := c.Bool("healthy")
	ttl := c.Int64("ttl")
	networks := makeLabels(c.StringSlice("network"))
	extension := c.String("extension")
	opts := &pb.SetContainersStatusOptions{Status: []*pb.ContainerStatus{}}
	for _, ID := range c.Args().Slice() {
		s := &pb.ContainerStatus{
			Id:        ID,
			Running:   running,
			Healthy:   healthy,
			Ttl:       ttl,
			Networks:  networks,
			Extension: []byte(extension),
		}
		opts.Status = append(opts.Status, s)
	}

	resp, err := client.SetContainersStatus(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	if c.Bool("pretty") {
		prettyRenderContainerStatus(resp.Status)
	} else {
		for _, containerStatus := range resp.Status {
			renderContainerStatus(containerStatus)
		}
	}
	return nil
}

func listContainers(c *cli.Context) error {
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()

	opts := &pb.ListContainersOptions{
		Appname:    c.Args().First(),
		Entrypoint: c.String("entry"),
		Nodename:   c.String("nodename"),
		Labels:     makeLabels(c.StringSlice("label")),
		Limit:      c.Int64("limit"),
	}

	resp, err := client.ListContainers(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	containers := []*pb.Container{}
	for {
		container, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.Exit(err, -1)
		}
		containers = append(containers, container)
	}

	if c.Bool("pretty") {
		prettyRenderContianers(containers)
	} else {
		for _, container := range containers {
			renderContainer(container)
		}
	}
	return nil
}

func reallocContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	memoryRequest, err := parseRAMInHuman(c.String("memory-request"))
	if err != nil {
		return cli.Exit(err, -1)
	}
	memoryLimit, err := parseRAMInHuman(c.String("memory-limit"))
	if err != nil {
		return cli.Exit(err, -1)
	}
	var volumesRequest, volumesLimit []string
	if v := c.String("volumes-request"); v != "" {
		volumesRequest = strings.Split(v, ",")
	}
	if v := c.String("volumes-limit"); v != "" {
		volumesLimit = strings.Split(v, ",")
	}
	bindCPU := c.Bool("cpu-bind")
	unbindCPU := c.Bool("cpu-unbind")

	if bindCPU && unbindCPU {
		return cli.Exit(errors.New("cpu-bind and cpu-unbind can not both be set"), -1)
	}
	bindCPUOpt := pb.TriOpt_KEEP
	if bindCPU {
		bindCPUOpt = pb.TriOpt_TRUE
	}
	if unbindCPU {
		bindCPUOpt = pb.TriOpt_FALSE
	}
	storageRequest, err := parseRAMInHuman(c.String("storage-request"))
	if err != nil {
		return cli.Exit(err, -1)
	}
	storageLimit, err := parseRAMInHuman(c.String("storage-limit"))
	if err != nil {
		return cli.Exit(err, -1)
	}

	opts := &pb.ReallocOptions{
		Id:         c.Args().First(),
		BindCpuOpt: bindCPUOpt,
		ResourceOpts: &pb.ResourceOptions{
			CpuQuotaRequest: c.Float64("cpu-request"),
			CpuQuotaLimit:   c.Float64("cpu-limit"),
			MemoryRequest:   memoryRequest,
			MemoryLimit:     memoryLimit,
			VolumesRequest:  volumesRequest,
			VolumesLimit:    volumesLimit,
			StorageRequest:  storageRequest,
			StorageLimit:    storageLimit,
		},
	}

	resp, err := client.ReallocResource(c.Context, opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	if resp.Error != "" {
		log.Infof("[Realloc] Failed by %+v", resp.Error)
	} else {
		log.Info("[Realloc] Success")
	}

	return nil
}

func execContainer(c *cli.Context) (err error) {
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()

	opts := &pb.ExecuteContainerOptions{
		ContainerId: c.Args().First(),
		OpenStdin:   c.Bool("interactive"),
		Commands:    c.Args().Tail(),
		Envs:        c.StringSlice("env"),
		Workdir:     c.String("workdir"),
	}
	resp, err := client.ExecuteContainer(context.Background())
	if err != nil {
		return
	}

	if err = resp.Send(opts); err != nil {
		return
	}

	iStream := interactiveStream{
		Recv: resp.Recv,
		Send: func(cmd []byte) error {
			return resp.Send(&pb.ExecuteContainerOptions{ReplCmd: cmd})
		},
	}

	code, err := handleInteractiveStream(opts.OpenStdin, iStream, 1)
	if err == nil {
		return cli.Exit("", code)
	}

	return err

}

func getContainerLog(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	tail := c.String("tail")

	opts := &pb.LogStreamOptions{Id: c.Args().First(), Tail: tail}
	resp, err := client.LogStream(c.Context, opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Error != "" {
			log.Errorf("[GetContainerLog] Failed %s %s", coreutils.ShortID(msg.Id), msg.Error)
			continue
		}

		log.Infof("[GetContainerLog] %s", string(msg.Data))
	}
	return nil
}

func copyContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	targets := map[string]*pb.CopyPaths{}
	for _, idSrc := range c.Args().Slice() {
		parts := strings.Split(idSrc, ":")
		paths := strings.Split(parts[1], ",")
		targets[parts[0]] = &pb.CopyPaths{Paths: paths}
	}

	opts := &pb.CopyOptions{Targets: targets}
	resp, err := client.Copy(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	now := time.Now().Format("2006.01.02.15.04.05")
	baseDir := filepath.Join(c.String("dir"))
	err = os.MkdirAll(baseDir, os.FileMode(0700)) // drwx------
	if err != nil {
		return cli.Exit(err, -1)
	}

	files := map[string]*os.File{}
	defer func() {
		// Close files
		for _, f := range files {
			f.Close()
		}
	}()
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Error != "" {
			log.Errorf("[Copy] Failed %s %s", coreutils.ShortID(msg.Id), msg.Error)
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s.tar.gz", coreutils.ShortID(msg.Id), msg.Name, now)
		storePath := filepath.Join(baseDir, filename)
		if _, err := os.Stat(storePath); err != nil {
			file, err := os.Create(storePath)
			if err != nil {
				log.Errorf("[Copy] Error during create backup file %s: %v", storePath, err)
				continue
			}
			files[storePath] = file
		}

		_, err = files[storePath].Write(msg.Data)
		if err != nil {
			log.Errorf("[Copy] Write file error %v", err)
		}
	}
	return nil
}

func sendContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	fileData := utils.GetFilesStream(c.StringSlice("file"))
	containerIDs := c.Args().Slice()
	opts := &pb.SendOptions{Ids: containerIDs, Data: fileData}
	resp, err := client.Send(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Error != "" {
			log.Errorf("[Send] Failed send %s to %s", msg.Path, msg.Id)
		} else {
			log.Infof("[Send] Send %s to %s success", msg.Path, msg.Id)
		}
	}

	return nil
}

func dissociateContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	containerIDs := c.Args().Slice()

	opts := &pb.DissociateContainerOptions{Ids: containerIDs}
	resp, err := client.DissociateContainer(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Error == "" {
			log.Infof("[Dissociate] Dissociate container %s from eru success", msg.Id)
		} else {
			log.Errorf("[Dissociate] Dissociate container %s from eru failed %v", msg.Id, msg.Error)
		}
	}

	return nil
}

func startContainers(c *cli.Context) error {
	return doControlContainers(c, cluster.ContainerStart)
}

func stopContainers(c *cli.Context) error {
	return doControlContainers(c, cluster.ContainerStop)
}

func restartContainers(c *cli.Context) error {
	return doControlContainers(c, cluster.ContainerRestart)
}

func doControlContainers(c *cli.Context, t string) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	opts := &pb.ControlContainerOptions{
		Ids: c.Args().Slice(), Type: t,
		Force: c.Bool("force"),
	}
	resp, err := client.ControlContainer(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		log.Infof("[ControlContainer] %s", coreutils.ShortID(msg.Id))
		if msg.Hook != nil {
			log.Infof("[ControlContainer] HookOutput %s", string(msg.Hook))
		}
		if msg.Error != "" {
			log.Errorf("Failed %s", msg.Error)
		}
	}
	return nil
}
