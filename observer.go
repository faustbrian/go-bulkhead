package bulkhead

import "time"

// EventKind identifies a bounded lifecycle transition.
type EventKind string

const (
	// EventAdmitted records committed capacity admission, before Acquire returns.
	EventAdmitted EventKind = "admitted"
	// EventQueued records queue insertion, not eventual admission.
	EventQueued EventKind = "queued"
	// EventRejected records a non-caller-cancellation admission failure.
	EventRejected EventKind = "rejected"
	// EventCanceled records rejected admission due to caller cancellation, not callback cancellation.
	EventCanceled EventKind = "canceled"
	// EventReleased records restored capacity, before release observation finishes.
	EventReleased EventKind = "released"
	// EventExecuted records finalized Execute timing, including callback errors/panics.
	EventExecuted EventKind = "executed"
	// EventDraining records the first transition to closed admission.
	EventDraining EventKind = "draining"
	// EventDrained records closed admission with no active weight or queue nodes.
	EventDrained EventKind = "drained"
)

// RejectionReason distinguishes terminal admission outcomes.
type RejectionReason string

const (
	// RejectionCapacity identifies immediate capacity saturation.
	RejectionCapacity RejectionReason = "capacity"
	// RejectionQueue identifies the configured queue-node bound.
	RejectionQueue RejectionReason = "queue_full"
	// RejectionTimeout identifies the maximum admission-wait deadline, not an execution timeout.
	RejectionTimeout RejectionReason = "wait_timeout"
	// RejectionCaller identifies nil or canceled acquiring contexts.
	RejectionCaller RejectionReason = "caller_canceled"
	// RejectionShutdown identifies closed admission.
	RejectionShutdown RejectionReason = "shutdown"
	// RejectionWeight identifies non-positive or over-capacity requested weight.
	RejectionWeight RejectionReason = "invalid_weight"
	// RejectionReentry identifies detectable same-instance Execute ancestry.
	RejectionReentry RejectionReason = "reentrant"
)

// Event contains copied lifecycle metadata, without contexts, callback results
// or raw errors. Resource and PolicyRevision are validated caller labels, not
// automatically redacted. Delivery can overlap/reorder across callers; state
// sampled in the event may already differ when an observer reads Snapshot.
type Event struct {
	// Resource is the configured caller-owned identity.
	Resource string
	// PolicyRevision is the optional configured caller-owned revision label.
	PolicyRevision string
	// Kind identifies the committed transition or recorded outcome.
	Kind EventKind
	// Reason is set for rejected/canceled admission, and empty otherwise.
	Reason RejectionReason
	// Weight is the requested/admitted weight; invalid requests can be non-positive
	// or over capacity. Lifecycle-only drain events use zero.
	Weight int64
	// QueueDepth counts retained queue nodes at the transition; drain-complete uses zero.
	QueueDepth int
	// ActiveWeight is committed reserved weight, not a count of running callbacks.
	ActiveWeight int64
	// WaitDuration is accounted wait for completed queued admission/rejection,
	// also carried by EventExecuted; other events use zero.
	WaitDuration time.Duration
	// ExecutionTime is callback elapsed time for EventExecuted, zero otherwise.
	ExecutionTime time.Duration
	// At is sampled at delivery, not the state linearization point. It may be
	// zero if the observer timestamp clock call panics.
	At time.Time
}

// Snapshot is a mutex-consistent state/lifetime-counter copy. RejectionCounts
// is a fresh caller-owned map; copies of a Snapshot share that map. Duration and
// counter aggregates are not reset or rolling windows. They use fixed-width Go
// arithmetic without saturation; the capacity bound does not bound lifetime sums.
type Snapshot struct {
	// Resource is the configured caller-owned identity.
	Resource string
	// PolicyRevision is the optional configured revision label.
	PolicyRevision string
	// Capacity is the fixed total weighted capacity.
	Capacity int64
	// ActiveWeight includes every committed grant, even before its caller receives a permit.
	ActiveWeight int64
	// AvailableWeight is Capacity minus ActiveWeight.
	AvailableWeight int64
	// QueueDepth counts retained nodes, including terminal nodes awaiting caller removal.
	QueueDepth int
	// Admissions counts committed immediate and queued grants.
	Admissions uint64
	// Rejections counts all failed Acquire outcomes, including caller cancellation.
	Rejections uint64
	// Cancellations counts acquiring-caller cancellation outcomes, not callback cancellations.
	Cancellations uint64
	// Executions counts finalized Execute callbacks, including errors/panics;
	// manually acquired/released permits do not increment it.
	Executions uint64
	// TotalWaitDuration sums accounted completed queue waits (admitted and failed),
	// excluding pending waits and immediate rejection/admission.
	TotalWaitDuration time.Duration
	// TotalExecutionDuration sums finalized Execute callback time before release.
	TotalExecutionDuration time.Duration
	// RejectionCounts is the copied per-reason failed-admission count map.
	RejectionCounts map[RejectionReason]uint64
	// Draining is true after admission is closed and remains true thereafter.
	Draining bool
	// Drained means admission is closed, active weight is zero and nodes are removed;
	// it does not certify observer completion or application goroutine termination.
	Drained bool
}
