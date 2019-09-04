package commands

import (
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

	"github.com/docker/go-units"
	"github.com/getlantern/deepcopy"
	"github.com/pkg/term/termios"
	pb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"bufio"

	enginecontainer "github.com/docker/docker/api/types/container"
)

type interactiveStream struct {
	Send func(cmd []byte) error
	Recv func() (*pb.AttachContainerMessage, error)
}

func makeLabels(labels []string) map[string]string {
	ext := map[string]string{}
	for _, d := range labels {
		if d == "" {
			continue
		}
		p := strings.Split(d, "=")
		if len(p) != 2 {
			continue
		}
		ext[p[0]] = p[1]
	}
	return ext
}

func getNetworks(network string) map[string]string {
	var ip string
	networkInfo := strings.Split(network, "=")
	if len(networkInfo) == 2 {
		network = networkInfo[0]
		ip = networkInfo[1]
	}
	networkmode := enginecontainer.NetworkMode(network)
	networks := map[string]string{}
	if network != "" && networkmode.IsUserDefined() {
		networks[network] = ip
	}
	return networks
}

func parseRAMInHuman(ramStr string) (int64, error) {
	if ramStr == "" {
		return 0, nil
	}
	flag := int64(1)
	if strings.HasPrefix(ramStr, "-") {
		flag = int64(-1)
		ramStr = strings.TrimLeft(ramStr, "-")
	}
	ramInBytes, err := units.RAMInBytes(ramStr)
	if err != nil {
		return 0, err
	}
	return ramInBytes * flag, nil
}

func handleInteractiveStream(interactive bool, iStream interactiveStream, exitCount int) (code int, err error) {

	if interactive {
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
					if err = iStream.Send(command); err != nil {
						log.Errorf("[handleInteractiveStream] Send SIGWINCH error: %v", err)
					}
				}
			}
		}(ctx)

		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(bufio.ScanRunes)
			for scanner.Scan() {
				b := scanner.Bytes()
				if err := iStream.Send(b); err != nil {
					log.Errorf("[handleInteractiveStream] Send command %s error: %v", b, err)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Errorf("[handleInteractiveStream] Failed to read output from virtual unit: %v", err)
				return
			}
		}()
	}

	exited := 0
	for {
		msg, err := iStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return -1, err
		}

		if bytes.HasPrefix(msg.Data, exitCode) {
			ret := string(bytes.TrimLeft(msg.Data, string(exitCode)))
			code, err = strconv.Atoi(ret)
			if err == nil && code != 0 {
				return code, err
			}
			exited++
			if exited == exitCount {
				return code, err
			}
			continue
		}

		if interactive {
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

	return
}
