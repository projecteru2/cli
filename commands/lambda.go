package commands

import (
	"context"
	"strings"

	"github.com/projecteru2/cli/utils"
	"github.com/projecteru2/core/cluster"
	pb "github.com/projecteru2/core/rpc/gen"
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
	client := setupAndGetGRPCConnection().GetRPCClient()
	code, err := lambda(c, client)
	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func lambda(c *cli.Context, client pb.CoreRPCClient) (code int, err error) {
	commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, count, stdin, deployMethods, files, user, async, asyncTimeout := getLambdaParams(c)
	opts := generateLambdaOpts(commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, count, stdin, deployMethods, files, user, async, asyncTimeout)

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
	pod string, envs []string, volumes []string,
	workingDir string, image string, cpu float64,
	mem int64, count int, stdin bool, deployMethod string,
	files []string, user string, async bool, asyncTimeout int) *pb.RunAndWaitOptions {

	networks := getNetworks(network)
	opts := &pb.RunAndWaitOptions{Async: async, AsyncTimeout: int32(asyncTimeout)}
	fileData := utils.GetFilesStream(files)
	opts.DeployOptions = &pb.DeployOptions{
		Name: "lambda",
		Entrypoint: &pb.EntrypointOptions{
			Name:       name,
			Command:    strings.Join(commands, " && "),
			Privileged: false,
			Dir:        workingDir,
		},
		Podname:      pod,
		Image:        image,
		CpuQuota:     cpu,
		Memory:       mem,
		Count:        int32(count),
		Env:          envs,
		Volumes:      volumes,
		Networks:     networks,
		Networkmode:  network,
		OpenStdin:    stdin,
		DeployMethod: deployMethod,
		Data:         fileData,
		User:         user,
	}
	return opts
}

func getLambdaParams(c *cli.Context) ([]string, string, string, string, []string, []string, string, string, float64, int64, int, bool, string, []string, string, bool, int) {
	if c.NArg() <= 0 {
		log.Fatal("[Lambda] no commands")
	}
	commands := c.Args().Slice()
	name := c.String("name")
	network := c.String("network")
	pod := c.String("pod")
	envs := c.StringSlice("env")
	volumes := c.StringSlice("volume")
	workingDir := c.String("working_dir")
	image := c.String("image")
	cpu := c.Float64("cpu")
	mem, err := parseRAMInHuman(c.String("memory"))
	if err != nil {
		log.Fatalf("[Lambda] memory wrong %v", err)
	}

	count := c.Int("count")
	stdin := c.Bool("stdin")
	files := c.StringSlice("file")
	deployMethod := c.String("deploy-method")
	user := c.String("user")
	async := c.Bool("async")
	asyncTimeout := c.Int("async-timeout")
	return commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, count, stdin, deployMethod, files, user, async, asyncTimeout
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
				Name:  "volume",
				Usage: "set volume can use multiple times",
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
				Name:  "cpu",
				Usage: "how many cpu",
				Value: 1.0,
			},
			&cli.StringFlag{
				Name:  "memory",
				Usage: "memory, support K, M, G, T",
				Value: "512M",
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
				Name:  "deploy-method",
				Usage: "deploy method auto/fill/each",
				Value: cluster.DeployAuto,
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
		},
		Action: runLambda,
	}
}
