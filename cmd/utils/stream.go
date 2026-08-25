package utils

import (
	"errors"
	"io"
)

// StreamToChan returns a wait func that reports the recv error once the channel is drained.
func StreamToChan[T any](recv func() (*T, error)) (<-chan *T, func() error) {
	var recvErr error
	ch := make(chan *T)
	go func() {
		defer close(ch)
		recvErr = EachMessage(recv, func(msg *T) error {
			ch <- msg
			return nil
		})
	}()
	return ch, func() error { return recvErr }
}

func EachMessage[T any](recv func() (*T, error), fn func(*T) error) error {
	var errs error
	for {
		msg, err := recv()
		if errors.Is(err, io.EOF) {
			return errs
		}
		if err != nil {
			return errors.Join(errs, err)
		}
		errs = errors.Join(errs, fn(msg))
	}
}
