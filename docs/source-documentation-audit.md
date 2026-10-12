# Source documentation audit

## Scope and acceptance

Tier A source/API documentation pass from
`c68a2c550e463903a537e4e4664804eb1983fa59`. The repository worker owns this
bounded change; runtime, dependencies, executable examples and existing work
remain unchanged. Named risks are weighted FIFO/admission race interpretation,
permit ownership, callback reentrancy/blocking, honest drain, metadata privacy
and registry pointer/snapshot lifetime. Coordinator owns delivery and shared
checker adoption, not this worker. Cataloged direct consumers are the two nested
harnesses and Service's `integration/adoption` and `integration/reference-external`
modules; their executable API/dependency inputs are unchanged.

## Reproducible inventory

```sh
git ls-files '*.go' 'go.mod' '**/go.mod'
cat modules.json packages.json .golib.yaml .go-version
rg -n '^package |^type |^func |^const |^var ' --glob '*.go'
rg -n 'go:generate|Code generated|go:build|\+build|TODO|FIXME|HACK|NOTE|WARNING|Deprecated:|nolint' --glob '*.go'
golib source-docs inventory --json
golib source-docs check --json
```

Three tracked modules: root `github.com/faustbrian/go-bulkhead` (public package
`bulkhead`), `benchmarks/comparison` (test-only `comparison_test` package) and
`integration/resilience` (test-only `resilience_test` package). The latter two
are non-releasable harnesses consuming published modules, not adapters or new
public production APIs; manifests deliberately declare no production packages
there. All three catalogs pin module Go 1.27.0; repository development pin is
1.27.2. There are 25 Go files: eight production and 17 test/fixture files.
No platform/build variants, generated Go files, generation directives,
deprecations or unfinished/warning markers are present. Existing directives
remain intact. Externally generated/vendor rewriting does not apply.

| Tracked Go file | Reviewed responsibility |
| --- | --- |
| `benchmark_test.go` | Admitted/saturated/parallel/wait/observer/partition measurements and sample sorting. |
| `benchmarks/comparison/benchmark_test.go` | Test-only nested comparison harness; capacity-one scenarios with different API overhead. |
| `boundaries_internal_test.go` | Validation and private waiter/drain accounting boundary fixtures. |
| `bulkhead.go` | Canonical package, semaphore/mutex ownership, admission/permit/counter state. |
| `bulkhead_test.go` | Immediate weighted admission and conservation. |
| `config.go` | Sealed policy choices, bounded labels/queue defaults, trusted clocks/observers. |
| `contract_test.go` | Weighted FIFO, absolute wait deadlines, cancellation, observers and custom clocks. |
| `drain.go` | Admission closure, queued wakeups and caller-bounded non-reclaiming drain. |
| `errors.go` | Stable admission/registry/cleanup sentinels and joined context identities. |
| `example_test.go` | Three unchanged output-bearing public execution/registry/drain examples. |
| `execution.go` | Synchronous callback, immutable context ancestry, deferred release and re-panic. |
| `execution_test.go` | Nil callbacks, admission rejection, callback errors/panics/reentry. |
| `fuzz_test.go` | Configuration boundaries and capacity-four permit histories truncated to 256 bytes. |
| `hardening_test.go` | Adversarial terminal paths, weighted starvation, observer reentry and partition removal. |
| `integration/resilience/budget_v2_test.go` | Test-only nested integration; Retry/Resilience v2 lineage, refusal and permit completion. |
| `integration/resilience/composition_test.go` | Application-owned local-rejection ordering before circuit-breaker admission and retry. |
| `kubernetes_model_test.go` | Readiness/drain/replacement model; not actual cluster/runtime proof. |
| `leak_test.go` | Goleak test-main boundary, no production background worker. |
| `model_test.go` | Randomized/concurrent histories and duplicate release. |
| `observer.go` | Lifecycle enums, event fields and caller-owned snapshot counters. |
| `registry.go` | Explicit bounded partition creation, borrowed pointers, lock order/removal snapshots. |
| `registry_test.go` | Isolation, cardinality, unknown keys and maximum validation. |
| `transition_race_test.go` | Grant/cancel/close/release lock-order transitions and single-terminal histories. |
| `wait.go` | FIFO list, absolute deadline, grant-before-cancel precedence and notification ownership. |
| `wait_test.go` | Bounded queue refusal, cancellation, timeout and release-triggered grants. |

## Gaps and truthful rationale

- The root package has only a one-line responsibility comment. Add construction,
  ownership/lifecycle, process-local non-goals and optional-infrastructure/cgo
  boundaries. Give both test-only harness packages a canonical scope comment.
- Public interface methods, Config/Wait/registry fields, lifecycle enums and
  Event/Result/Snapshot fields lack useful obligations or counter semantics.
  Inspect every public declaration/member, not only missing-comment counts.
- Bulkhead owns state/list/drained notification under its mutex; weighted
  semaphore accounting stays with x/sync v0.23.0. Permit atomic CAS reserves
  exactly one releaser, but concurrent duplicate calls need not wait for that
  first release's completion. Public instances require construction, not zero values.
- Queue insertion/grant/removal linearize under the policy mutex. Grant already
  committed wins even if cancellation/timer readiness is selected later.
  Cancellation is not continuously sampled under the granting lock. Absolute
  wait expiration checks `Now`, not timer delivery; queued observer latency uses
  the same budget. Expired/canceled nodes may retain queue slots until their
  callers remove them. Weighted head blocking prevents lighter bypass.
- Acquire returns an owned permit that survives cancellation; bare Acquire does
  not detect reentry without Execute's derived context. Execute invokes the
  callback synchronously exactly once after successful admission, forwards its
  value/error without disposal/retry, and releases/re-panics on callback panic.
  Separate policy instances with equal resource labels are not the same reentry
  identity. Callback contexts retain ancestry even after Execute returns; reusing
  one for later admission on an ancestor still rejects. Uncooperative work
  retains capacity until release/return.
- Observers run outside locks, may reenter Snapshot/Close, and can observe later
  state than the event. Their errors/panics are ignored, not their synchronous
  latency or successful reentrant side effects. Events have no ordering guarantee
  across callers. Clocks must be prompt/non-panicking; some Now sampling paths
  contain panic but initial admission/timer creation do not promise recovery.
- Close wakes queued callers without synchronously removing them or canceling
  active work. Drain watches queue removal and weight release, not application
  goroutine termination or completion of observer callbacks. A nil Drain context
  returns joined incomplete/canceled errors without closing admission; cancellation
  does not reclaim capacity. Concurrent ready/canceled selection is not a strict
  preference for either result.
- Snapshot's map is defensively copied and caller-mutable, not immutable in Go.
  Lifetime counters/durations use fixed-width arithmetic without saturation;
  a finite instantaneous capacity is not an unbounded overflow-free history.
  Registry snapshots use a copied membership list and per-policy snapshots, not
  one atomic cross-partition instant. Registry locking can acquire policy locking,
  while policy methods never acquire registry locking/call observers under locks.
  Removing a drained partition leaves borrowed pointers closed, allowing explicit
  replacement without reopening stale capacity.
- Resource/revision validation limits bytes/ASCII characters, not secrets or
  aggregate cardinality across standalone instances. Event.PolicyRevision is
  also caller metadata, contradicting the existing privacy wording. Limits bound
  queue/index counts, not memory retained by each caller/callback or global pods.

Local helpers with obvious syntax need no filler. Private capitalized methods
or fields are not public API merely because a syntax walker sees them. The
shared checker's known overreach is coordinator-owned; do not manufacture prose
solely to satisfy that proxy. No invariant requires a behavior repair in this
pass. Domain obligations (prompt callbacks, transport deadlines, safe labels,
per-pod sizing and application shutdown) remain explicit caller responsibilities.

## Verification plan

- Bounded source inventory/check, with any real structural findings reconciled
  separately from known checker overreach. Public comments and private rationale
  receive manual source/test review, not semantic certification by a count.
- Pinned formatting plus all-file code-token/comment-free AST equivalence and
  exact build/generation/linter directives; use isolated caches in one serialized
  session. Inspect rendered native root `go doc -all` and harness package comments.
- `.golib.yaml` selects apidiff `api/baseline.txt`; no repository-owned native
  snapshot generator or oracle exists in Makefile/scripts. Keep apidiff input
  unchanged. Native docs script renders root packages and validates unchanged
  examples; no example runtime change requires a new test run.
- Native Markdown link/anchor check and executable-example structure checks,
  generation/marker inventory, changed-content secret scan and one complete
  ordinary diff review. Update material public clarification in Unreleased.
- No broad runtime, nested dependency builds, race/leak, active fuzz, mutation,
  benchmark or clean-consumer campaign: executable contracts are preserved.
  Remote-main CI and any required checker adoption are coordinator-owned.

## Results

- Current shared `golib source-docs inventory --json` and `check --json` passed:
  three modules, 25 tracked Go files, no findings. Manual review covered all
  110 public declarations/members and the private/test responsibilities above.
  No boilerplate for private capitalized members was added to appease the checker.
- Catalog `GOTOOLCHAIN=go1.27.0`, `GOWORK=off`, bounded parallelism and three
  task-owned caches were used for one serialized local session. Pinned formatting
  passed for all 15 affected Go files. A disposable `go/scanner`/`go/parser`
  comparison against the recorded base passed for every Go file: code tokens,
  comment/source-position-free AST and compiler/linter directives match. Only
  optional inferred newline separators before closing delimiters are normalized.
- Complete native `go doc -all .` rendered stably twice, with all public contracts
  inspected. Stdlib `go/doc` rendered the canonical package comments of the two
  test-only harnesses without compiling their dependencies. The shared apidiff
  input is byte-unchanged; there is no repository-owned generated API snapshot.
- The native docs-script Markdown links/anchors and three unchanged examples'
  output/cleanup-error structure checks passed. All-file generation/build and
  complete warning/deprecation/unfinished-marker inventory passed.
- One ordinary complete-final-diff review, changed-content secret checks and
  ACK profile/message checks passed. Normal safe-wrapper hooks run at commit.
- No runtime/example execution, nested dependency build, race/leak, active fuzz,
  mutation, benchmark or clean-consumer campaign was run for this non-runtime
  pass. Existing fixtures informed prose; they are not newly passing proof.
- Exact owned Go caches/helper/rendered output were removed after capture. The
  committed worktree is retained for coordinator delivery. Shared checker CI
  adoption and exact-source remote-main qualification remain coordinator-owned.
- No unresolved technical question requires a behavioral change here. Callback
  cooperation, safe labels, bounded aggregate application state and shutdown
  remain caller responsibilities, not guarantees inferred from test counts.
