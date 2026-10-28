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

// StallWatch cancels one streamed request once it goes StreamIdleTimeout
// without data, from the moment it is sent until its body is read. A nil
// watch watches nothing, for a request that is not streamed.
type StallWatch struct {
	ctx   context.Context
	stop  context.CancelCauseFunc
	timer *time.Timer
}

// WatchStall starts a watch; send the request with its Context and Stop it
// when done.
func WatchStall(ctx context.Context) *StallWatch {
	w := &StallWatch{}
	w.ctx, w.stop = context.WithCancelCause(ctx)
	w.timer = time.AfterFunc(StreamIdleTimeout, func() { w.stop(ErrStalled) })
	return w
}

// Context is the request's context, cancelled when it stalls.
func (w *StallWatch) Context() context.Context { return w.ctx }

// Stop ends the watch and releases its context.
func (w *StallWatch) Stop() {
	w.timer.Stop()
	w.stop(nil)
}

// Err is err, or ErrStalled when the watch cancelled the request.
func (w *StallWatch) Err(err error) error {
	if w != nil && err != nil && errors.Is(context.Cause(w.ctx), ErrStalled) {
		return ErrStalled
	}
	return err
}

// Body counts every read of r as data, restarting the idle timeout.
func (w *StallWatch) Body(r io.Reader) io.Reader {
	if w == nil {
		return r
	}
	return stallReader{w, r}
}

type stallReader struct {
	w *StallWatch
	r io.Reader
}

func (s stallReader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	if n > 0 {
		s.w.timer.Reset(StreamIdleTimeout)
	}
	return n, s.w.Err(err)
}
