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
	exitCode     = []byte("[exitcode] ")
	winchCommand = []byte{0x80}
)

// Stream carries the send and recv half of an attach stream.
type Stream struct {
	Send func(cmd []byte) error
	Recv func() (*corepb.AttachWorkloadMessage, error)
}

type window struct {
	Row uint16
	Col uint16
}

// HandleStream pumps an attach stream, optionally putting the terminal in raw mode.
func HandleStream(ctx context.Context, interactive bool, iStream Stream, exitCount int, printWorkloadID bool) (int, error) {
	logger := log.WithFunc("interactive.HandleStream")

	if interactive {
		defer attachTerminal(ctx, iStream)()
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
			code, convErr = strconv.Atoi(string(bytes.TrimPrefix(msg.Data, exitCode)))
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

func attachTerminal(ctx context.Context, iStream Stream) func() {
	stdinFd := int(os.Stdin.Fd())
	ctx, cancel := context.WithCancel(ctx)

	state, err := term.MakeRaw(stdinFd)
	if err != nil {
		// stdin is a pipe or a file: no raw mode and no window size to report.
		go pumpStdin(ctx, iStream)
		return cancel
	}
	go pumpStdin(ctx, iStream)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	go watchWindowSize(ctx, iStream, stdinFd, sigs)

	return func() {
		cancel()
		signal.Stop(sigs)
		_ = term.Restore(stdinFd, state)
	}
}

func pumpStdin(ctx context.Context, iStream Stream) {
	logger := log.WithFunc("interactive.pumpStdin")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		if err := iStream.Send(scanner.Bytes()); err != nil {
			logger.Errorf(ctx, err, "send command %s", scanner.Bytes())
		}
	}
	logger.Error(ctx, scanner.Err(), "read stdin")
}

func watchWindowSize(ctx context.Context, iStream Stream, stdinFd int, sigs <-chan os.Signal) {
	logger := log.WithFunc("interactive.watchWindowSize")
	send := func() {
		col, row, err := term.GetSize(stdinFd)
		if err != nil {
			logger.Error(ctx, err, "get terminal size")
			return
		}
		opts, err := json.Marshal(&window{Row: uint16(row), Col: uint16(col)}) //nolint:gosec // terminal geometry fits in uint16
		if err != nil {
			logger.Error(ctx, err, "encode window size")
			return
		}
		logger.Error(ctx, iStream.Send(slices.Concat(winchCommand, opts)), "send window size")
	}

	send()
	for {
		select {
		case <-ctx.Done():
			return
		case <-sigs:
			send()
		}
	}
}
