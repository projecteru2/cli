package commands

import (
	"context"
	"strings"

	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/strategy"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var exitCode = []byte{91, 101, 120, 105, 116, 99, 111, 100, 101, 93, 32}
var enter = []byte{10}
var winchCommand = []byte{0x80}
var clrf = []byte{0xa}

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16 `json:"-"`
	Ypixel uint16 `json:"-"`
}

func runLambda(c *cli.Context) error {
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	code, err := lambda(c, client)
	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func lambda(c *cli.Context, client pb.CoreRPCClient) (code int, err error) {
	commands, name, network, pod, envs, volumesRequest, volumesLimit, workingDir, image, cpuRequest, cpuLimit, memRequest, memLimit, storageRequest, storageLimit, count, stdin, deployStrategys, files, user, async, asyncTimeout, priviledged, nodes := getLambdaParams(c)
	opts := generateLambdaOpts(commands, name, network, pod, envs, volumesRequest, volumesLimit, workingDir, image, cpuRequest, cpuLimit, memRequest, memLimit, storageRequest, storageLimit, count, stdin, deployStrategys, files, user, async, asyncTimeout, priviledged, nodes)

	resp, err := client.RunAndWait(context.Background())
	if err != nil {
		return -1, err
	}

	if resp.Send(opts) != nil {
		return -1, err
	}

	iStream := interactiveStream{
		Recv: resp.Recv,
		Send: func(cmd []byte) error {
			return resp.Send(&pb.RunAndWaitOptions{Cmd: cmd})
		},
	}

	go func() { _ = iStream.Send(clrf) }()
	return handleInteractiveStream(stdin, iStream, count)
}

func generateLambdaOpts(
	commands []string, name string, network string,
	pod string, envs []string, volumesRequest []string, volumesLimit []string,
	workingDir string, image string, cpuRequest, cpuLimit float64,
	memRequest, memLimit, storageRequest, storageLimit int64,
	count int, stdin bool, deployStrategy string,
	files []string, user string, async bool, asyncTimeout int, priviledged bool, nodes []string) *pb.RunAndWaitOptions {

	networks := getNetworks(network)
	opts := &pb.RunAndWaitOptions{Async: async, AsyncTimeout: int32(asyncTimeout)}
	fileData := utils.GetFilesStream(files)
	opts.DeployOptions = &pb.DeployOptions{
		Name: "lambda",
		Entrypoint: &pb.EntrypointOptions{
			Name:       name,
			Command:    strings.Join(commands, " && "),
			Privileged: priviledged,
			Dir:        workingDir,
		},
		ResourceOpts: &pb.ResourceOptions{
			CpuQuotaRequest: cpuRequest,
			CpuQuotaLimit:   cpuLimit,
			MemoryRequest:   memRequest,
			MemoryLimit:     memLimit,
			StorageRequest:  storageRequest,
			StorageLimit:    storageLimit,
			VolumesRequest:  volumesRequest,
			VolumesLimit:    volumesLimit,
		},
		Podname:        pod,
		Nodenames:      nodes,
		Image:          image,
		Count:          int32(count),
		Env:            envs,
		Networks:       networks,
		Networkmode:    network,
		OpenStdin:      stdin,
		DeployStrategy: pb.DeployOptions_Strategy(pb.DeployOptions_Strategy_value[strings.ToUpper(deployStrategy)]),
		Data:           fileData,
		User:           user,
	}
	return opts
}

func getLambdaParams(c *cli.Context) ([]string, string, string, string, []string, []string, []string, string, string, float64, float64, int64, int64, int64, int64, int, bool, string, []string, string, bool, int, bool, []string) {
	if c.NArg() <= 0 {
		log.Fatal("[Lambda] no commands")
	}
	commands := c.Args().Slice()
	name := c.String("name")
	network := c.String("network")
	pod := c.String("pod")
	envs := c.StringSlice("env")
	volumeRequest := c.StringSlice("volume-request")
	volumeLimit := c.StringSlice("volume")
	workingDir := c.String("working_dir")
	image := c.String("image")
	cpuRequest := c.Float64("cpu-request")
	cpuLimit := c.Float64("cpu")
	memRequest, err := parseRAMInHuman(c.String("memory-request"))
	if err != nil {
		log.Fatalf("[Lambda] memory wrong %v", err)
	}
	memLimit, err := parseRAMInHuman(c.String("memory"))
	if err != nil {
		log.Fatalf("[Lambda] memory wrong %v", err)
	}
	storageRequest := c.Int64("storage-request")
	storageLimit := c.Int64("storage")

	count := c.Int("count")
	stdin := c.Bool("stdin")
	files := c.StringSlice("file")
	deployStrategy := c.String("deploy-strategy")
	user := c.String("user")
	async := c.Bool("async")
	asyncTimeout := c.Int("async-timeout")
	privileged := c.Bool("privileged")
	nodes := c.StringSlice("node")
	return commands, name, network, pod, envs, volumeRequest, volumeLimit, workingDir, image, cpuRequest, cpuLimit, memRequest, memLimit, storageRequest, storageLimit, count, stdin, deployStrategy, files, user, async, asyncTimeout, privileged, nodes
}

// LambdaCommand for run commands in a container
func LambdaCommand() *cli.Command {
	return &cli.Command{
		Name:  "lambda",
		Usage: "run commands in a container like local",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "name for this lambda",
			},
			&cli.StringFlag{
				Name:  "network",
				Usage: "SDN name",
			},
			&cli.StringFlag{
				Name:  "pod",
				Usage: "where to run",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "set env can use multiple times, e.g., GO111MODULE=on",
			},
			&cli.StringSliceFlag{
				Name:  "volume-request",
				Usage: "set volume request can use multiple times",
			},
			&cli.StringSliceFlag{
				Name:  "volume",
				Usage: "set volume limitcan use multiple times",
			},
			&cli.StringFlag{
				Name:  "working_dir",
				Usage: "use as current working dir",
				Value: "/",
			},
			&cli.StringFlag{
				Name:  "image",
				Usage: "base image for running",
				Value: "alpine:latest",
			},
			&cli.Float64Flag{
				Name:  "cpu-request",
				Usage: "how many cpu request",
				Value: 0,
			},
			&cli.Float64Flag{
				Name:  "cpu",
				Usage: "how many cpu limit",
				Value: 1.0,
			},
			&cli.StringFlag{
				Name:  "memory-request",
				Usage: "memory request, support K, M, G, T",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "memory",
				Usage: "memory limit, support K, M, G, T",
				Value: "512M",
			},
			&cli.StringFlag{
				Name:  "storage-request",
				Usage: "how many storage to request quota like 1M or 1G, support K, M, G, T",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "storage",
				Usage: "how many storage to limit quota like 1M or 1G, support K, M, G, T",
				Value: "",
			},
			&cli.IntFlag{
				Name:  "count",
				Usage: "how many containers",
				Value: 1,
			},
			&cli.BoolFlag{
				Name:    "stdin",
				Usage:   "open stdin for container",
				Aliases: []string{"s"},
				Value:   false,
			},
			&cli.StringFlag{
				Name:  "deploy-strategy",
				Usage: "deploy method auto/fill/each",
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
			&cli.BoolFlag{
				Name:  "async",
				Usage: "run lambda async",
			},
			&cli.IntFlag{
				Name:  "async-timeout",
				Usage: "for async timeout",
				Value: 30,
			},
			&cli.BoolFlag{
				Name:    "privileged",
				Usage:   "give extended privileges to this lambda",
				Aliases: []string{"p"},
				Value:   false,
			},
			&cli.StringSliceFlag{
				Name:  "node",
				Usage: "which node to run",
			},
		},
		Action: runLambda,
	}
}
