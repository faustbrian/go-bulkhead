package bulkhead

import (
	"fmt"
	"reflect"
	"regexp"
	"time"
)

const (
	// MaxResourceBytes bounds resource/revision label bytes, not distinct label
	// counts or secret content.
	MaxResourceBytes = 128
	// MaxQueueSize bounds queued nodes per bulkhead, not caller-retained state.
	MaxQueueSize = 1 << 20
	// MaxPartitions bounds registered policies in one explicit registry.
	MaxPartitions = 1 << 20
)

var resourcePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]*$`)

// AdmissionPolicy selects immediate rejection or strict FIFO bounded waiting.
// New accepts only the concrete values RejectImmediately and Wait; nil selects
// immediate rejection. The private marker is not a custom-policy extension point.
type AdmissionPolicy interface{ admissionPolicy() }

// RejectImmediately never queues when capacity is unavailable.
type RejectImmediately struct{}

func (RejectImmediately) admissionPolicy() { _ = "reject-immediately" }

// Wait bounds the queue and supplies an absolute admission deadline. It does
// not forcibly interrupt observer callbacks or bound total Acquire wall time.
// Strict FIFO ordering can cause weighted head-of-line blocking.
type Wait struct {
	// MaxQueued is a positive node limit no greater than MaxQueueSize.
	// Terminal nodes retain their slot until their acquiring caller removes them.
	MaxQueued int
	// MaxWait is positive and starts at Acquire's initial clock sample.
	// Queued observer latency consumes this budget; an earlier committed grant
	// remains admitted even if notification is handled after the deadline.
	MaxWait time.Duration
}

func (Wait) admissionPolicy() { _ = "strict-fifo-wait" }

// Clock supplies timestamps and timers for waiting and observations. Shared
// implementations must be concurrency-safe, prompt and non-panicking. New uses
// the standard-library clock when Config.Clock is nil; typed nil is rejected.
type Clock interface {
	// Now supplies monotonic-comparable times for absolute deadlines and durations.
	Now() time.Time
	// NewTimer returns a non-nil live, stoppable timer for the remaining wait.
	NewTimer(time.Duration) Timer
}

// Timer is supplied by a Clock and owned by one waiting Acquire call.
// That caller stops it on return; the bulkhead never closes its channel.
type Timer interface {
	// C returns the usable timeout notification channel, without closure ownership.
	C() <-chan time.Time
	// Stop promptly releases timer resources and reports whether firing was prevented.
	Stop() bool
}

// Observer receives copied bounded lifecycle events synchronously. It runs
// outside internal locks; returned errors and panics do not reverse committed
// admission. Calls may overlap, and delivery order can differ from state order.
// Implementations must synchronize shared state and return promptly.
type Observer interface {
	// Observe receives metadata by value. Snapshot/Close reentry is allowed;
	// reentrant side effects still take effect. Waiting on the current call's
	// own admission/release (for example through Drain) can prevent that call
	// from completing. Blocking adds to admission, release and shutdown latency.
	Observe(Event) error
}

// ObserveFunc adapts a function to Observer.
type ObserveFunc func(Event) error

// Observe invokes the adapted function. Direct invocation of a nil function
// panics; New rejects typed nil observers.
func (function ObserveFunc) Observe(event Event) error { return function(event) }

// Config contains immutable construction parameters for one resource partition.
type Config struct {
	// Resource is non-empty, at most MaxResourceBytes bytes, begins with an ASCII
	// letter/digit and otherwise uses letters, digits or . _ : / -. The caller
	// must choose stable non-secret classes, not arbitrary request identifiers.
	Resource string
	// Capacity is positive weighted capacity, fixed for this policy's lifetime.
	Capacity int64
	// Admission defaults to immediate rejection when nil; use a concrete Wait
	// value for a bounded FIFO queue. Pointer/custom policy values are unsupported.
	Admission AdmissionPolicy
	// Observer is optional and borrowed; its state remains caller-owned.
	Observer Observer
	// Clock is borrowed, or nil for the standard clock. Typed nil is invalid.
	Clock Clock
	// PolicyRevision is optional and uses Resource's length/character bounds.
	// It labels a revision but does not reconfigure or replace the policy.
	PolicyRevision string
}

// normalizedConfig copies policy values but borrows observer/clock references;
// construction validates caller labels before storing them in events or indexes.
type normalizedConfig struct {
	resource       string
	capacity       int64
	wait           bool
	maxQueued      int
	maxWait        time.Duration
	observer       Observer
	clock          Clock
	policyRevision string
}

func normalize(config Config) (normalizedConfig, error) {
	if config.Resource == "" || len(config.Resource) > MaxResourceBytes || !resourcePattern.MatchString(config.Resource) {
		return normalizedConfig{}, invalidConfig("Resource", "must be a bounded metric-safe identifier")
	}
	if config.Capacity <= 0 {
		return normalizedConfig{}, invalidConfig("Capacity", "must be positive")
	}
	if len(config.PolicyRevision) > MaxResourceBytes ||
		(config.PolicyRevision != "" && !resourcePattern.MatchString(config.PolicyRevision)) {
		return normalizedConfig{}, invalidConfig("PolicyRevision", "must be empty or a bounded metric-safe identifier")
	}
	if nilLike(config.Clock) || nilLike(config.Observer) || nilLike(config.Admission) {
		return normalizedConfig{}, invalidConfig("dependency", "must not be a typed nil")
	}

	normalized := normalizedConfig{
		resource:       config.Resource,
		capacity:       config.Capacity,
		observer:       config.Observer,
		clock:          config.Clock,
		policyRevision: config.PolicyRevision,
	}
	if normalized.clock == nil {
		normalized.clock = systemClock{}
	}

	switch policy := config.Admission.(type) {
	case nil, RejectImmediately:
	case Wait:
		if policy.MaxQueued <= 0 || policy.MaxQueued > MaxQueueSize {
			return normalizedConfig{}, invalidConfig("Admission.MaxQueued", "must be positive and bounded")
		}
		if policy.MaxWait <= 0 {
			return normalizedConfig{}, invalidConfig("Admission.MaxWait", "must be positive")
		}
		normalized.wait = true
		normalized.maxQueued = policy.MaxQueued
		normalized.maxWait = policy.MaxWait
	default:
		return normalizedConfig{}, invalidConfig("Admission", "unsupported policy")
	}

	return normalized, nil
}

func invalidConfig(field, message string) error {
	return fmt.Errorf("%w: %s %s", ErrInvalidConfig, field, message)
}

func nilLike(value any) bool {
	if value == nil {
		return false
	}
	kind := reflect.ValueOf(value).Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) &&
		reflect.ValueOf(value).IsNil()
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func (systemClock) NewTimer(duration time.Duration) Timer {
	return systemTimer{Timer: time.NewTimer(duration)}
}

type systemTimer struct{ *time.Timer }

func (timer systemTimer) C() <-chan time.Time { return timer.Timer.C }
