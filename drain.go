package bulkhead

import (
	"context"
	"errors"
)

// Close idempotently stops new admission and wakes pending queued callers with
// ErrClosed. Already committed grants/terminal outcomes remain authoritative.
// Existing permits retain capacity and must be released; protected work is not
// canceled. Close does not wait for queued callers to remove their nodes or for
// active work to finish. Observers run synchronously and can delay its return.
func (bulkhead *Bulkhead) Close() error {
	bulkhead.mu.Lock()
	if bulkhead.draining {
		bulkhead.mu.Unlock()
		return nil
	}
	bulkhead.draining = true
	for element := bulkhead.waiters.Front(); element != nil; element = element.Next() {
		waiter := element.Value.(*waiter)
		if waiter.done {
			continue
		}
		waiter.done = true
		waiter.terminal = ErrClosed
		close(waiter.ready)
	}
	event := bulkhead.eventLocked(EventDraining, "", 0, 0, 0)
	drained := bulkhead.maybeDrainedLocked()
	bulkhead.mu.Unlock()

	bulkhead.observe(event)
	bulkhead.observeDrained(drained)
	return nil
}

// Drain closes admission and waits for zero active weight and removed queued
// nodes, not application goroutine/observer completion. Concurrent calls are safe.
// Context termination returns ErrDrainIncomplete joined with ctx.Err without
// reclaiming live capacity. A nil context returns an incomplete/canceled error
// without closing admission. If completion and cancellation are both ready,
// either select result may occur. Close's synchronous observer is not forcibly
// interrupted by ctx; callbacks must cooperate for bounded shutdown latency.
func (bulkhead *Bulkhead) Drain(ctx context.Context) error {
	if ctx == nil {
		return errors.Join(ErrDrainIncomplete, context.Canceled)
	}
	_ = bulkhead.Close()
	select {
	case <-bulkhead.drained:
		return nil
	case <-ctx.Done():
		return errors.Join(ErrDrainIncomplete, ctx.Err())
	}
}
