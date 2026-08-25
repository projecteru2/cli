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
	}, nil)

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
