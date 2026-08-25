package interactive

import (
	"sync"
	"testing"
)

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
