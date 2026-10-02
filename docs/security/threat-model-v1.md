# Security model v1

Model version: 1. Reviewed on 2026-09-30 against root module source
`1224c26e9737f50f584a5d2cfbcd3e25ed88e6f9` (Go 1.27), public release
`v1.0.0`. This model documents existing behavior; it does not certify all
ecosystem consumers or a future release.

## Scope and trust boundaries

The root package owns process-local weighted admission, explicit partition
registration, queued waits, permit release, and drain accounting. The nested
`integration/resilience` module exercises composition. Direct ecosystem
consumers include `go-service` adoption and external-reference integrations,
and the maintained `go-library-tools` fixture. No consumer update is required
for this documentation-only audit.

An untrusted request may influence weight, admission frequency, cancellation,
and application-selected partition lookup. Configuration, partition creation,
observers, clocks, and operation callbacks are trusted application inputs.
Applications must not delegate unrestricted policy creation to requests.

Protected assets are capacity availability, bounded queue and registry memory,
correct lifecycle accounting, and non-secret operational metadata. The package
performs no network, filesystem, process, environment, authentication,
authorization, credential storage, or cryptographic work. It is not a tenant
authorization boundary, distributed semaphore, retry engine, or deadline
enforcer for protected work.

## Boundary and evidence matrix

| Boundary or attack | Existing control | Behavioral evidence |
| --- | --- | --- |
| Label injection and oversized metadata | Resource and policy revision accept bounded ASCII identifiers, at most 128 bytes; invalid configuration does not echo values | `contract_test.go`: `TestConfigurationRejectsUnboundedAndTypedNilInputs`, `TestConfigurationAcceptsDocumentedUpperBounds` |
| Partition cardinality and churn | Explicit creation with a fixed maximum; lookup never creates; removal requires fully drained state | `registry_test.go` |
| Pathological weights and capacity overflow | Positive weight no greater than configured capacity; permits release once; accounting cannot exceed capacity | `contract_test.go`, `hardening_test.go`, `model_test.go` |
| Queue exhaustion and unfair bypass | Mandatory positive bounded queue and wait; strict weighted FIFO, no lighter bypass | `wait_test.go`, `TestMixedWeightFIFOHeadCannotBeStarvedByLighterArrivals` |
| Cancellation, grant, expiry, and shutdown races | Exactly one terminal admission; a reserved grant wins a racing cancellation; Close preserves live permits; Drain is context bounded | `transition_race_test.go`, `execution_test.go`, `kubernetes_model_test.go` |
| Callback error or panic | Observer errors/panics ignored outside locks; operation errors returned; operation panics rethrown after release | `TestObserverFailurePanicAndReentrancyDoNotAlterAdmission`, `TestEveryAdmittedTerminalPathReleasesExactlyOnePermit` |
| Secret-bearing observability | Events exclude operation results, contexts, and errors; validation errors omit rejected values | `observer.go`, `config.go`; application label secrecy remains caller-owned |
| Supply-chain compromise | Module checksums and immutable shared-workflow pin; private disclosure entry point | `go.sum`, `.github/workflows/ci.yml`, [reporting policy](../../SECURITY.md) |

### Reused immutable verification

[CI run 36551565037](https://github.com/faustbrian/go-bulkhead/actions/runs/36551565037)
on 2026-09-29 passed at the exact audited source revision. Its root module
contract executed vet, race, lint, Staticcheck, vulnerability, secrets,
licenses, fuzz, docs, API, and NilAway gates. Logs report no vulnerabilities
and no secret leaks; both configuration and permit-history fuzz targets ran
100 executions and passed. NilAway remains warning-only, not an enforced
absence-of-nil-defects claim. CodeQL completed successfully. Dependency review
and release rehearsal were skipped in this scheduled run; they are not claimed
as passing. The ordinary module contract covers the named tests above, without
claiming that each was run in isolation.

| Module | Exact-source CI outcome | Security disposition | Release verdict |
| --- | --- | --- | --- |
| Root `github.com/faustbrian/go-bulkhead` | Root quality contract and Required passed in run 36551565037 | No confirmed critical/high source finding; caller-owned risks below accepted conditionally | Documentation only; retain v1.0.0, no module release |
| `integration/resilience` | Its selected quality contract passed in the same run | Composition test module; no production boundary changed | Non-releasable test harness; no release |
| `benchmarks/comparison` | Its selected quality contract passed in the same run | Benchmark-only dependencies isolated; not runtime security assurance | Non-releasable benchmark harness; no release |

Reuse applies to unchanged source, dependencies, and pinned automation. New
documentation still requires its own link/structure checks and delivered-head
CI. These scanner results do not establish callback liveness, application
policy correctness, undisclosed vulnerability absence, or security of direct
external consumers.

## Accepted residual risks

Repository disposition owner: `faustbrian`, the package maintainer. Operational
mitigation owner: the deploying application's maintainer, identified by that
application's ownership policy. Medium denotes possible service-availability
loss requiring trusted application configuration or callback behavior, not a
confirmed remote authorization bypass. Low denotes caller-controlled exposure
without a package-owned secret source. The package maintainer owns reassessment
when a listed review condition changes.

| Risk | Owner and rationale | Mitigation | Review condition |
| --- | --- | --- | --- |
| Medium: a synchronous observer blocks or recursively invokes unbounded callbacks | Application maintainer; availability can fail, but exploitation requires application-controlled code rather than request input alone; arbitrary callbacks cannot be forcibly stopped safely | Use a prompt concurrency-safe observer; bound any caller-owned handoff and avoid admission recursion; package maintainer documents synchronous ownership | New asynchronous observer contract, incident involving callback stalls, or observer implementation change |
| Medium: custom Clock or Timer blocks, panics, returns nil, or does not progress | Application maintainer; these are trusted injected timing collaborators, not hostile extensions; invalid implementations can strand waits | Prefer the default clock; custom implementations must be prompt, non-panicking, concurrency-safe and return live stoppable timers | Any new clock adapter, production clock injection, or timer/lifecycle change |
| Medium: protected operation ignores cancellation or caller retains a permit | Application maintainer; reclaiming capacity while work still runs would break the isolation invariant | Release permits on every path or use Execute; apply transport deadlines and audit context propagation; supervise bounded shutdown | New protected integration, cancellation incident, or permit ownership change |
| Medium: configuration permits large memory budgets or aggregate replica amplification | Application maintainer; hard ceilings are safety bounds rather than recommended budgets | Use small explicit partition/queue budgets, limit request fan-out, retries and hedges, and account for replicas | Capacity policy, replica count, retry/hedge policy, or tenant mapping change |
| Low: syntactically valid labels contain secrets or high-cardinality identifiers | Application maintainer; identifier grammar cannot determine business confidentiality | Use fixed non-secret resource classes and revision labels; never raw request/customer/token values | New label derivation or observability exporter |
| Low: application errors or original operation panic values expose data | Application maintainer; the package preserves caller error/panic semantics and never publishes them in events | Redact at the application error/panic boundary before logging or responding | New callback or application logging/recovery policy |

No critical or high in-package finding was confirmed by this source audit.
Accepted risks are conditional on the mitigations above, not a claim that all
applications implement them. Dependency vulnerabilities, maintainer compromise,
and malicious releases require current scanning, review, private disclosure,
and release integrity controls in addition to this design model.

## Release and maintenance verdict

This audit changes documentation only: no public API, behavior, dependency,
module version, consumer source lock, or release is required. Keep the latest
stable v1 support policy in [SECURITY.md](../../SECURITY.md). Confirmed future
vulnerabilities require affected-version identification, focused regressions,
changelog and upgrade guidance, and coordinated disclosure; do not infer an
advisory from an accepted trusted-callback limitation.

Review this model after any admission, queue, registry, lifecycle, callback,
observability, dependency, or build/release trust-boundary change. Revise the
model version when its assumptions or risk disposition materially change.
