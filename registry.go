package bulkhead

import (
	"slices"
	"strings"
	"sync"
)

// PartitioningPolicy explicitly names how a registry bounds resource identities.
// NewRegistry accepts only a concrete FixedPartitions value; nil, pointers and
// custom implementations are not supported extension points.
type PartitioningPolicy interface{ partitioningPolicy() }

// FixedPartitions allows at most Maximum explicitly created partitions. It
// never creates from lookup keys and never evicts automatically.
type FixedPartitions struct {
	// Maximum is positive and no greater than MaxPartitions. Registration grows
	// incrementally; construction does not allocate this many partitions.
	Maximum int
}

func (FixedPartitions) partitioningPolicy() { _ = "fixed-partitions" }

// Registry owns an application-scoped bounded set of independent bulkheads.
// It is not global; applications choose and share its lifetime explicitly.
// Methods are concurrency-safe. Use NewRegistry rather than a nil/zero registry,
// and do not copy it. Returned policies are borrowed pointers; callers own
// closing/draining them, and removing them never reopens retained pointers.
type Registry struct {
	// mu owns membership. Remove may acquire a partition mutex while holding it;
	// partitions never lock this registry or invoke observers under either lock.
	mu         sync.RWMutex
	maximum    int
	partitions map[string]*Bulkhead
}

// NewRegistry validates a concrete FixedPartitions cardinality from one through
// MaxPartitions. Invalid or absent policies wrap ErrInvalidConfig. It allocates
// an empty index and starts no worker; the application owns partition shutdown.
func NewRegistry(policy PartitioningPolicy) (*Registry, error) {
	if nilLike(policy) {
		return nil, invalidConfig("PartitioningPolicy", "must not be a typed nil")
	}
	fixed, ok := policy.(FixedPartitions)
	if !ok {
		return nil, invalidConfig("PartitioningPolicy", "unsupported policy")
	}
	if fixed.Maximum <= 0 || fixed.Maximum > MaxPartitions {
		return nil, invalidConfig("FixedPartitions.Maximum", "must be positive and bounded")
	}
	return &Registry{
		maximum:    fixed.Maximum,
		partitions: make(map[string]*Bulkhead),
	}, nil
}

// Create validates and explicitly registers one immutable resource policy.
// It returns the shared stored pointer. Config errors precede ErrPartitionExists
// for an existing resource and ErrPartitionLimit for a full index. Registration
// failures leave membership unchanged.
func (registry *Registry) Create(config Config) (*Bulkhead, error) {
	partition, err := New(config)
	if err != nil {
		return nil, err
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.partitions[config.Resource]; exists {
		return nil, ErrPartitionExists
	}
	if len(registry.partitions) >= registry.maximum {
		return nil, ErrPartitionLimit
	}
	registry.partitions[config.Resource] = partition
	return partition, nil
}

// Lookup returns the borrowed registered pointer or ErrPartitionNotFound.
// It neither validates nor creates from a key. A concurrent Remove can unlink
// the pointer after lookup; the retained closed policy remains a valid instance.
func (registry *Registry) Lookup(resource string) (*Bulkhead, error) {
	registry.mu.RLock()
	partition, exists := registry.partitions[resource]
	registry.mu.RUnlock()
	if !exists {
		return nil, ErrPartitionNotFound
	}
	return partition, nil
}

// Remove deletes only a closed, fully drained partition. Configuration changes
// therefore require Close, Drain, Remove, and an explicit Create; retained
// pointers remain closed and cannot split resource capacity. Unknown resources
// yield ErrPartitionNotFound; open/active/still-queued policies yield ErrPartitionBusy.
func (registry *Registry) Remove(resource string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	partition, exists := registry.partitions[resource]
	if !exists {
		return ErrPartitionNotFound
	}
	if snapshot := partition.Snapshot(); !snapshot.Drained {
		return ErrPartitionBusy
	}
	delete(registry.partitions, resource)
	return nil
}

// Len reports a concurrent snapshot of current registered partition count.
func (registry *Registry) Len() int {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return len(registry.partitions)
}

// Snapshots copies membership, then returns resource-sorted per-policy snapshots
// with independent rejection-count maps. It is not one atomic cross-partition
// view; concurrent creation/removal and policy activity may proceed between samples.
func (registry *Registry) Snapshots() []Snapshot {
	registry.mu.RLock()
	partitions := make([]*Bulkhead, 0, len(registry.partitions))
	for _, partition := range registry.partitions {
		partitions = append(partitions, partition)
	}
	registry.mu.RUnlock()

	snapshots := make([]Snapshot, 0, len(partitions))
	for _, partition := range partitions {
		snapshots = append(snapshots, partition.Snapshot())
	}
	slices.SortFunc(snapshots, func(left, right Snapshot) int {
		return strings.Compare(left.Resource, right.Resource)
	})
	return snapshots
}
