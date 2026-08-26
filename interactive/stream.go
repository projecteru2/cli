package interactive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"sync"
	"syscall"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/term"
)

var (
	exitCode     = []byte("[exitcode] ")
	winchCommand = []byte{0x80}
)

type messageWriter func(io.Writer, *corepb.AttachWorkloadMessage) error

type window struct {
	Row uint16
	Col uint16
}

// Stream carries the send and recv half of an attach stream.
type Stream struct {
	Send      func(cmd []byte) error
	Recv      func() (*corepb.AttachWorkloadMessage, error)
	CloseSend func() error
}

// NewStream serializes send and closeSend, which grpc forbids calling from several goroutines.
func NewStream(send func(cmd []byte) error, recv func() (*corepb.AttachWorkloadMessage, error), closeSend func() error) Stream {
	var mu sync.Mutex
	closed := false
	return Stream{
		Send: func(cmd []byte) error {
			mu.Lock()
			defer mu.Unlock()
			if closed {
				return nil
			}
			return send(cmd)
		},
		Recv: recv,
		CloseSend: func() error {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			return closeSend()
		},
	}
}

// HandleStream pumps an attach stream, optionally putting the terminal in raw mode.
func HandleStream(ctx context.Context, interactive bool, iStream Stream, exitCount int, printWorkloadID bool) (int, error) {
	logger := log.WithFunc("interactive.HandleStream")

	if interactive {
		defer attachTerminal(ctx, iStream)()
	}

	write := outputWriter(printWorkloadID)
	code, exited := 0, 0
	for {
		msg, err := iStream.Recv()
		if errors.Is(err, io.EOF) {
			if exitCount == 0 {
				return code, nil
			}
			return -1, fmt.Errorf("stream ended after %d of %d workloads exited", exited, exitCount)
		}
		if err != nil {
			return -1, err
		}

		switch {
		case msg.StdStreamType == corepb.StdStreamType_ERUERROR:
			return -1, errors.New(string(msg.Data))
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
		if err := write(outStream, msg); err != nil {
			logger.Error(ctx, err, "write output")
		}
	}
}

func outputWriter(printWorkloadID bool) messageWriter {
	if printWorkloadID {
		return func(w io.Writer, msg *corepb.AttachWorkloadMessage) error {
			_, err := fmt.Fprintf(w, "[%s] %s", msg.WorkloadId, msg.Data)
			return err
		}
	}
	return func(w io.Writer, msg *corepb.AttachWorkloadMessage) error {
		_, err := w.Write(msg.Data)
		return err
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
		log.WithFunc("interactive.attachTerminal").Error(ctx, term.Restore(stdinFd, state), "restore terminal")
	}
}

func pumpStdin(ctx context.Context, iStream Stream) {
	logger := log.WithFunc("interactive.pumpStdin")
	buf := make([]byte, 32*1024)
	for {
		n, err := os.Stdin.Read(buf)
		if ctx.Err() != nil {
			return
		}
		if n > 0 {
			if sendErr := iStream.Send(buf[:n]); sendErr != nil {
				logger.Errorf(ctx, sendErr, "send %d bytes", n)
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				logger.Error(ctx, err, "read stdin")
			}
			logger.Error(ctx, iStream.CloseSend(), "close the send side")
			return
		}
	}
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
