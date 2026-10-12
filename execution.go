package bulkhead

import (
	"context"
	"errors"
	"reflect"
	"time"
)

var errAdmissionInvariant = errors.New("bulkhead admission invariant violated")

type executionContextKey struct{}

// executionScope is immutable ancestry attached only to the callback context.
// Identity is the Bulkhead pointer, not the Resource label; dropping this context
// loses detection, and manually held Acquire permits do not add ancestry.
type executionScope struct {
	bulkhead *Bulkhead
	parent   *executionScope
}

func (scope *executionScope) contains(target *Bulkhead) bool {
	for current := scope; current != nil; current = current.parent {
		if current.bulkhead == target {
			return true
		}
	}
	return false
}

// Result reports queue and execution duration separately for an admitted call.
type Result struct {
	// Resource is the admitted policy's caller-provided identity.
	Resource string
	// Weight is the admitted weight, not a count of downstream calls.
	Weight int64
	// WaitDuration is accounted queue wait; immediate admission reports zero.
	WaitDuration time.Duration
	// ExecutionDuration is non-negative callback elapsed time measured before
	// deferred release/observer delivery, not total Execute wall time.
	ExecutionDuration time.Duration
}

// Execute admits operation, invokes it with same-instance reentrancy detection,
// and releases capacity on return or panic. Cancellation cannot terminate an
// operation that ignores its context; such an operation retains capacity until
// it actually returns.
//
// A nil operation yields ErrInvalidOperation before admission; admission failure
// returns zero value/Result without invoking it. An admitted callback runs once,
// synchronously, with derived context values/cancellation intact. Its value/error
// are returned unchanged, including partial values; no disposal, retry or transport
// timeout is supplied. Detection requires keeping the derived context and compares
// policy instances, not equal resource labels. Ancestry is not removed from a
// retained callback context after return; do not reuse it for new admission on
// an ancestor policy. The Bulkhead must come from New.
func Execute[T any](
	ctx context.Context,
	bulkhead *Bulkhead,
	weight int64,
	operation func(context.Context) (T, error),
) (value T, result Result, err error) {
	switch reflect.ValueOf(operation).IsNil() {
	case true:
		return value, result, ErrInvalidOperation
	}
	permit, err := bulkhead.Acquire(ctx, weight)
	switch err {
	case nil:
	default:
		return value, result, err
	}
	return executeAdmitted(ctx, bulkhead, weight, permit, operation)
}

func executeAdmitted[T any](
	ctx context.Context,
	bulkhead *Bulkhead,
	weight int64,
	permit *Permit,
	operation func(context.Context) (T, error),
) (value T, result Result, err error) {
	if permit == nil {
		return value, result, errAdmissionInvariant
	}
	result = Result{Resource: bulkhead.config.resource, Weight: weight, WaitDuration: permit.waitDuration}
	executionStart := bulkhead.now()

	// Release even on callback panic, then preserve its original panic value.
	// Execution timing ends before release and its synchronous observation.
	defer func() {
		recovered := recover()
		result.ExecutionDuration = nonNegativeDuration(bulkhead.now().Sub(executionStart))
		_ = permit.Release()
		bulkhead.recordExecution(result)
		if recovered != nil {
			panic(recovered)
		}
	}()

	parent, _ := ctx.Value(executionContextKey{}).(*executionScope)
	executionContext := context.WithValue(ctx, executionContextKey{}, &executionScope{
		bulkhead: bulkhead,
		parent:   parent,
	})
	value, err = operation(executionContext)
	return value, result, err
}

func (bulkhead *Bulkhead) recordExecution(result Result) {
	bulkhead.mu.Lock()
	bulkhead.executions++
	bulkhead.totalExecution += result.ExecutionDuration
	event := bulkhead.eventLocked(
		EventExecuted, "", result.Weight, result.WaitDuration, result.ExecutionDuration,
	)
	bulkhead.mu.Unlock()
	bulkhead.observe(event)
}
