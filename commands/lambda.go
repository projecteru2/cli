package commands

import (
	"C"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/getlantern/deepcopy"
	"github.com/pkg/term/termios"
	"github.com/projecteru2/cli/utils"
	"github.com/projecteru2/core/cluster"
	pb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

var exitCode = []byte{91, 101, 120, 105, 116, 99, 111, 100, 101, 93, 32}
var enter = []byte{10}
var split = []byte{62, 32}
var winchCommand = []byte{0xf, 0xa}

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16 `json:"-"`
	Ypixel uint16 `json:"-"`
}

func runLambda(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()
	code, err := lambda(c, client)
	if err != nil {
		return cli.Exit(err, code)
	}
	return nil
}

func lambda(c *cli.Context, client pb.CoreRPCClient) (code int, err error) {
	commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, count, stdin, deployMethods, files, user := getLambdaParams(c)
	opts := generateLambdaOpts(commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, count, stdin, deployMethods, files, user)

	resp, err := client.RunAndWait(context.Background())
	if err != nil {
		return -1, err
	}

	if resp.Send(opts) != nil {
		return -1, err
	}

	if stdin {
		stdinFd := os.Stdin.Fd()
		terminal := &syscall.Termios{}
		termios.Tcgetattr(stdinFd, terminal)
		terminalBak := &syscall.Termios{}
		deepcopy.Copy(terminalBak, terminal)
		defer termios.Tcsetattr(stdinFd, termios.TCSANOW, terminalBak)

		// turn off echoing in terminal
		terminal.Lflag &^= syscall.ECHO
		termios.Tcsetattr(stdinFd, termios.TCSAFLUSH, terminal)

		// set uncanonical mode
		terminal.Lflag &^= syscall.ICANON
		termios.Tcsetattr(stdinFd, termios.TCSAFLUSH, terminal)

		// suppress terminal special characters
		suppressSpecials := []uint8{
			syscall.VINTR,   // ^C
			syscall.VEOF,    // ^D
			syscall.VSUSP,   // ^Z
			syscall.VKILL,   // ^U
			syscall.VERASE,  // ^?
			syscall.VWERASE, // ^W
		}
		for _, s := range suppressSpecials {
			terminal.Cc[s] = 0
		}
		termios.Tcsetattr(stdinFd, termios.TCSAFLUSH, terminal)

		// capture SIGWINCH and measure window size
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGWINCH)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func(ctx context.Context) {
			w := &window{}
			for {
				select {
				case <-ctx.Done():
					break
				case _, ok := <-sigs:
					if !ok {
						return
					}

					if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, stdinFd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(w))); err != 0 {
						return
					}
					opts, err := json.Marshal(w)
					if err != nil {
						return
					}
					command := append(winchCommand, opts...)
					if err = resp.Send(&pb.RunAndWaitOptions{Cmd: command}); err != nil {
						log.Errorf("[Lambda] Send SIGWINCH error: %v", err)
					}
				}
			}
		}(ctx)

		go func() {
			// 获得输入
			buf := make([]byte, 1024)
			for {
				n, err := os.Stdin.Read(buf)
				if n > 0 {
					command := buf[:n]
					if err = resp.Send(&pb.RunAndWaitOptions{Cmd: command}); err != nil {
						log.Errorf("[Lambda] Send command %s error: %v", command, err)
					}
				}
				if err != nil {
					if err == io.EOF {
						return
					}
					log.Errorf("[runAndWait] failed to read output from virtual unit: %v", err)
					return
				}
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

		if stdin {
			fmt.Printf("%s", msg.Data)
		} else {
			data := msg.Data
			id := coreutils.ShortID(msg.ContainerId)
			if !bytes.HasSuffix(data, split) {
				data = append(data, enter...)
			}
			fmt.Printf("[%s]: %s", id, data)
		}
	}
	return 0, nil
}

func generateLambdaOpts(
	commands []string, name string, network string,
	pod string, envs []string, volumes []string,
	workingDir string, image string, cpu float64,
	mem int64, count int, stdin bool, deployMethod string,
	files []string, user string) *pb.RunAndWaitOptions {

	networks := getNetworks(network)
	opts := &pb.RunAndWaitOptions{}
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

func getLambdaParams(c *cli.Context) ([]string, string, string, string, []string, []string, string, string, float64, int64, int, bool, string, []string, string) {
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
	return commands, name, network, pod, envs, volumes, workingDir, image, cpu, mem, count, stdin, deployMethod, files, user
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
		},
		Action: runLambda,
	}
}
