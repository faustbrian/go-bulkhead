// Package comparison_test is a non-releasable benchmark harness for capacity-one
// admitted/saturated Bulkhead, x/sync, Failsafe and Fortify APIs. Candidates have
// different lifecycle/queue/error contracts; results are not interchangeable
// production guarantees. This test-only module uses published dependencies and
// adds no optional infrastructure or cgo bindings to the public Bulkhead package.
// Run benchmarks from this module with workspace resolution disabled.
package comparison_test

import (
	"context"
	"testing"

	failsafebulkhead "github.com/failsafe-go/failsafe-go/bulkhead"
	"github.com/faustbrian/go-bulkhead"
	fortifybulkhead "go.klarlabs.de/fortify/bulkhead"
	"golang.org/x/sync/semaphore"
)

// BenchmarkAdmittedFastPath excludes policy/context setup and measures
// capacity-one acquire/release for native permit APIs, versus a no-op Execute
// callback for Fortify. These paths include different lifecycle/observation work.
func BenchmarkAdmittedFastPath(b *testing.B) {
	b.Run("Bulkhead", func(b *testing.B) {
		policy := mustPolicy(b)
		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			permit, err := policy.Acquire(ctx, 1)
			if err != nil {
				b.Fatalf("Acquire() error = %v", err)
			}
			_ = permit.Release()
		}
	})

	b.Run("XSyncSemaphore", func(b *testing.B) {
		policy := semaphore.NewWeighted(1)
		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if err := policy.Acquire(ctx, 1); err != nil {
				b.Fatalf("Acquire() error = %v", err)
			}
			policy.Release(1)
		}
	})

	b.Run("FailsafeGo", func(b *testing.B) {
		policy := failsafebulkhead.New[any](1)
		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if err := policy.AcquirePermit(ctx); err != nil {
				b.Fatalf("AcquirePermit() error = %v", err)
			}
			policy.ReleasePermit()
		}
	})

	b.Run("Fortify", func(b *testing.B) {
		policy := fortifybulkhead.New[struct{}](fortifybulkhead.Config{MaxConcurrent: 1})
		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_, err := policy.Execute(ctx, func(context.Context) (struct{}, error) { return struct{}{}, nil })
			if err != nil {
				b.Fatalf("Execute() error = %v", err)
			}
		}
	})
}

// BenchmarkRejectedFastPath holds all capacity before timing immediate refusal.
// The Fortify holder goroutine and final cleanup are outside timing; its refusal
// uses Execute, unlike the permit/error or boolean APIs of the other candidates.
func BenchmarkRejectedFastPath(b *testing.B) {
	b.Run("Bulkhead", func(b *testing.B) {
		policy := mustPolicy(b)
		holder, err := policy.Acquire(context.Background(), 1)
		if err != nil {
			b.Fatalf("holder Acquire() error = %v", err)
		}
		defer func() { _ = holder.Release() }()
		ctx := context.Background()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_, _ = policy.Acquire(ctx, 1)
		}
	})

	b.Run("XSyncTryAcquire", func(b *testing.B) {
		policy := semaphore.NewWeighted(1)
		if !policy.TryAcquire(1) {
			b.Fatal("holder TryAcquire() failed")
		}
		defer policy.Release(1)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if policy.TryAcquire(1) {
				policy.Release(1)
				b.Fatal("saturated TryAcquire() succeeded")
			}
		}
	})

	b.Run("FailsafeGo", func(b *testing.B) {
		policy := failsafebulkhead.New[any](1)
		if !policy.TryAcquirePermit() {
			b.Fatal("holder TryAcquirePermit() failed")
		}
		defer policy.ReleasePermit()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if policy.TryAcquirePermit() {
				policy.ReleasePermit()
				b.Fatal("saturated TryAcquirePermit() succeeded")
			}
		}
	})

	b.Run("Fortify", func(b *testing.B) {
		policy := fortifybulkhead.New[struct{}](fortifybulkhead.Config{MaxConcurrent: 1})
		started := make(chan struct{})
		finish := make(chan struct{})
		completed := make(chan error, 1)
		go func() {
			_, err := policy.Execute(context.Background(), func(context.Context) (struct{}, error) {
				close(started)
				<-finish
				return struct{}{}, nil
			})
			completed <- err
		}()
		<-started
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_, _ = policy.Execute(context.Background(), func(context.Context) (struct{}, error) {
				b.Fatal("saturated Fortify bulkhead invoked protected work")
				return struct{}{}, nil
			})
		}
		b.StopTimer()
		close(finish)
		if err := <-completed; err != nil {
			b.Fatalf("holder Execute() error = %v", err)
		}
	})
}

func mustPolicy(b *testing.B) *bulkhead.Bulkhead {
	b.Helper()
	policy, err := bulkhead.New(bulkhead.Config{Resource: "benchmark", Capacity: 1})
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	return policy
}
