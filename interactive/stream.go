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
	"slices"
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

// Stream carries the send and recv half of an attach stream.
type Stream struct {
	Send func(cmd []byte) error
	Recv func() (*corepb.AttachWorkloadMessage, error)
}

// HandleStream pumps an attach stream, optionally putting the terminal in raw mode.
func HandleStream(ctx context.Context, interactive bool, iStream Stream, exitCount int, printWorkloadID bool) (int, error) {
	logger := log.WithFunc("interactive.HandleStream")

	if interactive {
		detach, err := attachTerminal(ctx, iStream)
		if err != nil {
			return -1, err
		}
		defer detach()
	}

	outputTemplate := `{{printf "%s" .Data}}`
	if printWorkloadID {
		outputTemplate = `[{{.WorkloadId}}] {{printf "%s" .Data}}`
	}

	outputT, err := template.New("output").Parse(outputTemplate)
	if err != nil {
		return -1, err
	}

	code, exited := 0, 0
	for {
		msg, err := iStream.Recv()
		if errors.Is(err, io.EOF) {
			return code, nil
		}
		if err != nil {
			return -1, err
		}

		switch {
		case msg.StdStreamType == corepb.StdStreamType_ERUERROR:
			logger.Error(ctx, errors.New(string(msg.Data)), "error from eru")
			continue
		case msg.StdStreamType == corepb.StdStreamType_TYPEWORKLOADID:
			logger.Infof(ctx, "workload id %s", msg.WorkloadId)
			continue
		case bytes.HasPrefix(msg.Data, exitCode):
			var convErr error
			code, convErr = strconv.Atoi(string(bytes.TrimLeft(msg.Data, string(exitCode))))
			if convErr == nil && code != 0 {
				return code, nil
			}
			exited++
			if exited == exitCount {
				return code, convErr
			}
			continue
		}

		outStream := os.Stderr
		if msg.StdStreamType == corepb.StdStreamType_STDOUT {
			outStream = os.Stdout
		}
		if err := outputT.Execute(outStream, msg); err != nil {
			logger.Error(ctx, err, "render template")
		}
	}
}

// attachTerminal puts stdin in raw mode and forwards keystrokes and resizes to the stream.
func attachTerminal(ctx context.Context, iStream Stream) (func(), error) {
	logger := log.WithFunc("interactive.attachTerminal")
	stdinFd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(stdinFd)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)

	resize := func() error {
		col, row, err := term.GetSize(stdinFd)
		if err != nil {
			return err
		}
		opts, err := json.Marshal(&window{Row: uint16(row), Col: uint16(col)}) //nolint:gosec // terminal geometry fits in uint16
		if err != nil {
			return err
		}
		return iStream.Send(slices.Concat(winchCommand, opts))
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			case <-sigs:
				if err := resize(); err != nil {
					logger.Error(ctx, err, "resize")
				}
			}
		}
	}()

	go func() {
		if err := resize(); err != nil {
			logger.Error(ctx, err, "resize")
		}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(bufio.ScanRunes)
		for scanner.Scan() {
			if err := iStream.Send(scanner.Bytes()); err != nil {
				logger.Errorf(ctx, err, "send command %s", scanner.Bytes())
			}
		}
		if err := scanner.Err(); err != nil {
			logger.Error(ctx, err, "read stdin")
		}
	}()

	return func() {
		cancel()
		signal.Stop(sigs)
		_ = term.Restore(stdinFd, state)
	}, nil
}
