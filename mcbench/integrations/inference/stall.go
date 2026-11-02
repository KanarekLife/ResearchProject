package inference

import (
	"context"
	"errors"
	"io"
	"time"
)

// StreamIdleTimeout is how long a streamed request may receive nothing before
// it is abandoned as stalled. A gateway can hold a request for minutes without
// a byte while a retry answers in seconds; a stream that is working sends
// tokens or keep-alives well within this. It is a variable only so tests need
// not wait.
var StreamIdleTimeout = 2 * time.Minute

// ErrStalled is a streamed request that received nothing for
// StreamIdleTimeout. It is retried like a network failure.
var ErrStalled = errors.New("stream stalled: no data within the idle timeout")

// stallWatch cancels one streamed request once it goes StreamIdleTimeout
// without data, from the moment it is sent until its body is read.
type stallWatch struct {
	ctx   context.Context
	stop  context.CancelCauseFunc
	timer *time.Timer
}

// watchStall starts a watch; send the request with its context and stop it
// when done.
func watchStall(ctx context.Context) *stallWatch {
	w := &stallWatch{}
	w.ctx, w.stop = context.WithCancelCause(ctx)
	w.timer = time.AfterFunc(StreamIdleTimeout, func() { w.stop(ErrStalled) })
	return w
}

// end ends the watch and releases its context.
func (w *stallWatch) end() {
	w.timer.Stop()
	w.stop(nil)
}

// err is err, or ErrStalled when the watch cancelled the request.
func (w *stallWatch) err(err error) error {
	if err != nil && errors.Is(context.Cause(w.ctx), ErrStalled) {
		return ErrStalled
	}
	return err
}

// body counts every read of r as data, restarting the idle timeout.
func (w *stallWatch) body(r io.Reader) io.Reader { return stallReader{w, r} }

type stallReader struct {
	w *stallWatch
	r io.Reader
}

func (s stallReader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	if n > 0 {
		s.w.timer.Reset(StreamIdleTimeout)
	}
	return n, s.w.err(err)
}
