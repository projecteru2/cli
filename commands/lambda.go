package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	enginecontainer "github.com/docker/docker/api/types/container"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

var exitCode = []byte{91, 101, 120, 105, 116, 99, 111, 100, 101, 93, 32}
var enter = []byte{10}
var split = []byte{62, 32}

func runLambda(c *cli.Context) error {
	conn := setupAndGetGRPCConnection()
	code, err := lambda(c, conn)
	if err != nil {
		return cli.Exit(err, code)
	}
	return nil
}

func lambda(c *cli.Context, conn *grpc.ClientConn) (code int, err error) {
	commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, timeout, count, stdin := getLambdaParams(c)
	opts := generateLambdaOpts(commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, timeout, count, stdin)

	client := pb.NewCoreRPCClient(conn)
	resp, err := client.RunAndWait(context.Background())
	if err != nil {
		return -1, err
	}

	if resp.Send(opts) != nil {
		return -1, err
	}

	if stdin {
		go func() {
			// 获得输入
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				command := scanner.Bytes()
				log.Debugf("input: %s", command)
				command = append(command, enter...)
				if err = resp.Send(&pb.RunAndWaitOptions{Cmd: command}); err != nil {
					log.Errorf("[Lambda] Send command %s error: %v", command, err)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Errorf("[Lambda] Parse stdin failed, %v", err)
			}
		}()
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return -1, err
		}

		if bytes.HasPrefix(msg.Data, exitCode) {
			ret := string(bytes.TrimLeft(msg.Data, string(exitCode)))
			code, err = strconv.Atoi(ret)
			if err != nil {
				return code, err
			}
			continue
		}
		data := msg.Data
		id := msg.ContainerId[:7]
		if !bytes.HasSuffix(data, split) {
			data = append(data, enter...)
		}
		fmt.Printf("[%s]: %s", id, data)
	}
	return 0, nil
}

func generateLambdaOpts(
	commands []string, name string, network string,
	pod string, envs []string, volumes []string,
	workingDir string, image string, cpu float64,
	mem int64, timeout, count int, stdin bool) *pb.RunAndWaitOptions {

	networks := map[string]string{}
	if network != "" {
		networkmode := enginecontainer.NetworkMode(network)
		if networkmode.IsUserDefined() {
			networks = map[string]string{network: ""}
		}
	}
	opts := &pb.RunAndWaitOptions{}
	opts.DeployOptions = &pb.DeployOptions{
		Name: "lambda",
		Entrypoint: &pb.EntrypointOptions{
			Name:       name,
			Command:    strings.Join(commands, " && "),
			Privileged: false,
			Dir:        workingDir,
		},
		Podname:     pod,
		Image:       image,
		CpuQuota:    cpu,
		Memory:      mem,
		Count:       int32(count),
		Env:         envs,
		Volumes:     volumes,
		Networks:    networks,
		Networkmode: network,
		OpenStdin:   stdin,
	}
	opts.Timeout = int32(timeout)
	return opts
}

func getLambdaParams(c *cli.Context) ([]string, string, string, string, []string, []string, string, string, float64, int64, int, int, bool) {
	if c.NArg() <= 0 {
		log.Fatal("[Lambda] no commands ")
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
	mem := c.Int64("mem")
	timeout := c.Int("timeout")
	count := c.Int("count")
	stdin := c.Bool("stdin")
	return commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, timeout, count, stdin
}

//LambdaCommand for run commands in a container
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
				Usage: "set env can use multiple times",
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
			&cli.Int64Flag{
				Name:  "mem",
				Usage: "how many memory in bytes",
				Value: 536870912,
			},
			&cli.IntFlag{
				Name:  "timeout",
				Usage: "when to interrupt",
				Value: 60,
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
		},
		Action: runLambda,
	}
}
