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
		_ = termios.Tcgetattr(stdinFd, terminal)
		terminalBak := &syscall.Termios{}
		_ = deepcopy.Copy(terminalBak, terminal)
		defer func() { _ = termios.Tcsetattr(stdinFd, termios.TCSANOW, terminalBak) }()

		terminal.Lflag &^= syscall.ECHO   // off echoing
		terminal.Lflag &^= syscall.ICANON // noncanonical mode
		terminal.Lflag &^= syscall.ISIG   // disable signals
		terminal.Lflag &^= syscall.IEXTEN // extended input processing

		terminal.Iflag &^= syscall.BRKINT // disable special handling of BREAK
		terminal.Iflag &^= syscall.ICRNL  // disable special handling of CR
		terminal.Iflag &^= syscall.IGNBRK //disable special handling of BREAK
		terminal.Iflag &^= syscall.IGNCR  // disable special handling of CR
		terminal.Iflag &^= syscall.INLCR  // disable special handling of NL
		terminal.Iflag &^= syscall.INPCK  // no parity error handling
		terminal.Iflag &^= syscall.ISTRIP // no 8th-bit stripping
		terminal.Iflag &^= syscall.IXON   // disable output flow control
		terminal.Iflag &^= syscall.PARMRK // no parity error handling

		terminal.Oflag &^= syscall.OPOST // disable all output processing

		terminal.Cc[syscall.VMIN] = 1  // character-at-a-time input
		terminal.Cc[syscall.VTIME] = 0 // blocking

		_ = termios.Tcsetattr(stdinFd, termios.TCSAFLUSH, terminal)

		// capture SIGWINCH and measure window size
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGWINCH)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		resize := func(ctx context.Context) error {
			w := &window{}
			if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, stdinFd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(w))); err != 0 {
				return err
			}
			opts, err := json.Marshal(w)
			if err != nil {
				return err
			}
			command := append(winchCommand, opts...)
			return iStream.Send(command)
		}

		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					break
				case _, ok := <-sigs:
					if !ok {
						return
					}
					if err := resize(ctx); err != nil {
						log.Errorf("[handleInteractiveStream] Resize error: %v", err)
					}
				}
			}
		}(ctx)

		go func() {
			if err := resize(ctx); err != nil {
				log.Errorf("[handleInteractiveStream] Resize error: %v", err)
			}
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
	output := map[string][]byte{}
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
			id := coreutils.ShortID(msg.ContainerId)
			if _, ok := output[id]; !ok {
				output[id] = []byte{}
			}

			output[id] = append(output[id], msg.Data...)

			if bytes.HasSuffix(output[id], enter) {
				fmt.Printf("[%s]: %s", id, output[id])
				output[id] = []byte{}
			}
		}
	}

	for id, o := range output {
		if len(o) > 0 {
			fmt.Printf("[%s]: %s", id, output[id])
		}
	}

	return
}
