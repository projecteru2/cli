package interactive

import (
	"io"
	"strings"
	"sync"
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestHandleStreamRequiresTerminalStatus(t *testing.T) {
	tests := []struct {
		name      string
		msgs      []*corepb.AttachWorkloadMessage
		exitCount int
		wantCode  int
		wantErr   string
	}{
		{name: "synchronous eof", exitCount: 1, wantCode: -1, wantErr: "0 of 1 workloads exited"},
		{name: "core error", msgs: []*corepb.AttachWorkloadMessage{{StdStreamType: corepb.StdStreamType_ERUERROR, Data: []byte("allocate failed")}}, exitCount: 1, wantCode: -1, wantErr: "allocate failed"},
		{name: "asynchronous eof", wantCode: 0},
		{name: "exit code", msgs: []*corepb.AttachWorkloadMessage{{Data: []byte("[exitcode] 0")}}, exitCount: 1, wantCode: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := 0
			stream := Stream{Recv: func() (*corepb.AttachWorkloadMessage, error) {
				if next == len(tt.msgs) {
					return nil, io.EOF
				}
				msg := tt.msgs[next]
				next++
				return msg, nil
			}}

			code, err := HandleStream(t.Context(), false, stream, tt.exitCount, false)
			if code != tt.wantCode {
				t.Errorf("code: got %d, want %d", code, tt.wantCode)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("got %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewStreamSerializesSend(t *testing.T) {
	inFlight, peak := 0, 0
	stream := NewStream(func([]byte) error {
		inFlight++
		peak = max(peak, inFlight)
		inFlight--
		return nil
	}, nil, func() error { return nil })

	wg := sync.WaitGroup{}
	for range 64 {
		wg.Go(func() {
			if err := stream.Send([]byte("x")); err != nil {
				t.Errorf("send: %v", err)
			}
		})
	}
	wg.Wait()

	if peak != 1 {
		t.Errorf("got %d concurrent sends, want 1", peak)
	}
}

func TestPumpStdinClosesTheSendSideOnEOF(t *testing.T) {
	closed := false
	stream := NewStream(func([]byte) error { return nil }, nil, func() error {
		closed = true
		return nil
	})

	pumpStdin(t.Context(), stream)

	if !closed {
		t.Error("stdin EOF did not close the send side")
	}
}

func TestSendAfterCloseSendIsANoOp(t *testing.T) {
	sent := 0
	stream := NewStream(func([]byte) error { sent++; return nil }, nil, func() error { return nil })

	if err := stream.CloseSend(); err != nil {
		t.Fatalf("close send: %v", err)
	}
	if err := stream.Send([]byte("late")); err != nil {
		t.Fatalf("late send: %v", err)
	}
	if sent != 0 {
		t.Errorf("a send after CloseSend reached the wire %d times", sent)
	}
}
