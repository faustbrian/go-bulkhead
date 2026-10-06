package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faustbrian/go-bulkhead"
	resilience "github.com/faustbrian/go-resilience/v2"
	"github.com/faustbrian/go-retry/v2"
)

func TestBulkheadRetryVersion2BudgetAdmissionAndRefusal(t *testing.T) {
	bulkheadPolicy, err := bulkhead.New(bulkhead.Config{Resource: "dependency", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	budget, err := resilience.NewBudget(resilience.BudgetConfig{
		MaxResources: 1, MaxScopes: 1, MaxAdditionalPerExecution: 1,
		MaxConcurrentAdditional: 1, MaxAdditionalPerWindow: 1,
		AdditionalWindow: time.Minute, PermitTTL: time.Minute, Clock: retry.SystemClock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := resilience.NewMetadata("logical", "lookup", "dependency")
	if err != nil {
		t.Fatal(err)
	}
	scope, ctx, err := budget.Start(context.Background(), metadata)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := scope.Close(); err != nil {
			t.Errorf("close scope: %v", err)
		}
	})
	policy, err := retry.NewPolicyStrict(retry.Config{
		Backoff: retry.Constant(0), MaxAttempts: 3,
		Clock: retry.SystemClock{}, Sleeper: retry.SystemSleeper{},
		Classifier: retry.RetryableClassifier(), UseResilienceBudget: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var calls, protectedCalls uint64
	result, err := retry.DoStrict(ctx, policy, func(ctx context.Context) (retry.AttemptResult[struct{}], error) {
		calls++
		attempt, ok := resilience.AttemptFromContext(ctx)
		if !ok || attempt.Ordinal != calls {
			t.Fatal("missing attempt lineage")
		}
		if calls == 1 && (attempt.Origin != resilience.OriginOriginal || attempt.ParentOrdinal != 0) {
			t.Fatal("wrong original lineage")
		}
		if calls == 2 && (attempt.Origin != resilience.OriginRetry || attempt.ParentOrdinal != 1) {
			t.Fatal("wrong retry lineage")
		}
		value, _, executeErr := bulkhead.Execute(ctx, bulkheadPolicy, 1, func(context.Context) (struct{}, error) {
			protectedCalls++
			return struct{}{}, retry.Retryable(errors.New("temporary failure"))
		})
		return retry.AttemptResult[struct{}]{Value: value, Outcome: retry.OutcomeKnown}, executeErr
	})
	var rejection *resilience.BudgetRejectionError
	if calls != 2 || result.Outcome != retry.OutcomeKnown || result.Retry.Attempts != 2 ||
		result.Retry.Reason != retry.ReasonWorkBudget || !errors.Is(err, resilience.ErrBudgetRejected) ||
		!errors.As(err, &rejection) || rejection.Reason != resilience.ReasonExecutionLimit {
		t.Fatalf("budget refusal: calls=%d result=%+v error=%v", calls, result, err)
	}
	if snapshot := scope.Snapshot(); snapshot.AdditionalAdmitted != 1 || snapshot.AdditionalActive != 0 {
		t.Fatalf("scope snapshot=%+v", snapshot)
	}
	if protectedCalls != 2 {
		t.Fatalf("protected calls=%d, want 2", protectedCalls)
	}
	permit, err := bulkheadPolicy.Acquire(context.Background(), 1)
	if err != nil {
		t.Fatalf("retry leaked bulkhead capacity: %v", err)
	}
	if err := permit.Release(); err != nil {
		t.Fatal(err)
	}
}
