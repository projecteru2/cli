package interactive

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/term"
)

var (
	exitCode     = []byte{91, 101, 120, 105, 116, 99, 111, 100, 101, 93, 32}
	winchCommand = []byte{0x80}
)

type window struct {
	Row uint16
	Col uint16
}

// Stream is a wrapper for send and recv method
type Stream struct {
	Send func(cmd []byte) error
	Recv func() (*corepb.AttachWorkloadMessage, error)
}

// HandleStream will handle a stream with send and recv method
// with or without interactive mode
func HandleStream(ctx context.Context, interactive bool, iStream Stream, exitCount int, printWorkloadID bool) (code int, err error) {
	logger := log.WithFunc("interactive.HandleStream")

	if interactive {
		stdinFd := int(os.Stdin.Fd())
		state, err := term.MakeRaw(stdinFd)
		if err != nil {
			return -1, err
		}
		defer func() { _ = term.Restore(stdinFd, state) }()

		// capture SIGWINCH and measure window size
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGWINCH)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		resize := func() error {
			col, row, err := term.GetSize(stdinFd)
			if err != nil {
				return err
			}
			opts, err := json.Marshal(&window{Row: uint16(row), Col: uint16(col)})
			if err != nil {
				return err
			}
			command := append(winchCommand, opts...) //nolint
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
					if err := resize(); err != nil {
						logger.Error(ctx, err, "resize")
					}
				}
			}
		}(ctx)

		go func() {
			if err := resize(); err != nil {
				logger.Error(ctx, err, "resize")
			}
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(bufio.ScanRunes)
			for scanner.Scan() {
				b := scanner.Bytes()
				if err := iStream.Send(b); err != nil {
					logger.Errorf(ctx, err, "send command %s", b)
				}
			}
			if err := scanner.Err(); err != nil {
				logger.Error(ctx, err, "read output from virtual unit")
				return
			}
		}()
	}

	outputTemplate := `{{printf "%s" .Data}}`
	if printWorkloadID {
		outputTemplate = `[{{.WorkloadId}}] {{printf "%s" .Data}}`
	}

	outputT, err := template.New("output").Parse(outputTemplate)
	if err != nil {
		return -1, err
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

		// error should be printed and skipped
		if msg.StdStreamType == corepb.StdStreamType_ERUERROR {
			logger.Error(ctx, errors.New(string(msg.Data)), "error from eru")
			continue
		}

		if msg.StdStreamType == corepb.StdStreamType_TYPEWORKLOADID {
			logger.Infof(ctx, "workload id %s", msg.WorkloadId)
			continue
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

		var outStream *os.File
		switch msg.StdStreamType {
		case corepb.StdStreamType_STDOUT:
			outStream = os.Stdout
		default:
			outStream = os.Stderr
		}
		if err := outputT.Execute(outStream, msg); err != nil {
			logger.Error(ctx, err, "render template")
		}
	}

	return code, err
}
