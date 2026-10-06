# Changelog

All notable changes to this module are documented here.

## Unreleased

## 1.1.1 - 2026-10-06

### Changed

- Adopt published Retry v2.1 and Resilience v2.0 in the non-releasable resilience
  integration harness, preserving existing composition behavior and asserting
  attached-scope retry lineage, bounded admission, refusal, and permit completion.

- Refresh the documentation TOML parser lock and pinned library tooling
  while preserving bulkhead APIs, permit behavior, and the Go support floor.

## 1.1.0 - 2026-10-02

### Changed

- Raise the minimum supported Go version from 1.26.6 to 1.27.0. Consumers
  must upgrade their toolchain to use this release; the public API and
  runtime behavior remain unchanged.

- Update the runtime semaphore dependency to x/sync v0.23.0 while retaining
  the existing capacity validation, admission, and permit behavior.

- Update the comparison harness to Failsafe-Go v0.9.7.

- Update the comparison harness to Fortify v1.10.0 and select patched gRPC
  v1.83.2 for its dependency graph, retaining the public bulkhead API and
  runtime behavior while clearing the introduced security advisories.

- Upgrade the non-releasable resilience composition to Retry v1.1.0 and its
  strict policy, execution, and known-outcome contract so local bulkhead
  rejection remains conclusively non-retryable without implying an ambiguous
  downstream side effect.

- Adopt the checksum-verified `go-library-tools` v1.4.0 CLI and immutable W14
  reusable workflow, and reconcile nested Golib dependency checksums with
  their published v1.0.0 archives without changing selected versions, the
  bulkhead API, or runtime behavior.

- Adopt the `go-library-tools` v1.3.0 schema-v2 cohesion contract and local
  `make cohesion` gate without changing the bulkhead API or runtime behavior.
- Pin reusable CI to the v1.3.0 workflow and enforce cohesion metadata in the
  repository's required CI contract.

- Replace copied repository tooling with the pinned `go-library-tools` v1.0.13
  contract while retaining package-owned policy and verification evidence.

### Documentation

- Document the public package and internal harness map, add direct saturation
  troubleshooting, bind package-owned documentation verification, handle
  public cleanup errors in examples, clarify module tag conventions, route
  support and vulnerability reports, and correct the v1.0.0 release date.

- Add canonical v1 installation, stable Go support, lifecycle and ownership,
  project support, and security-reporting guidance.

- Link ecosystem and Resilience family guidance to the immutable v1.4.0
  documentation release.

- Publish the module's family, capabilities, ownership, lifecycle, supported
  environments, package selection, and delivery status, and link the README to
  the immutable v1.3.0 ecosystem index and family guidance.

- Keep the README focused on adoption and move detailed assurance guidance to
  the package documentation index.

## 1.0.0 - 2026-08-26

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-bulkhead` identity while preserving its documented API and behavior.

### Documentation

- Link the package README to the repository-wide Golib documentation portal.

### Added

- Fixed-capacity and weighted process-local bulkheads with stable resource
  identity, immediate rejection, strict FIFO bounded waiting, typed terminal
  admission outcomes, and exactly-once owned permits.
- Generic context-aware execution with separate wait and execution timing,
  panic-safe release, detectable same-policy reentrancy, and honest behavior
  for callbacks that ignore cancellation.
- Bounded explicit partition registries, immutable policy revisions,
  synchronous failure-contained observations, snapshots, and graceful
  application-driven drain.
- Kubernetes sizing and shutdown guidance, resilience composition contracts,
  operations, migration, security, FAQ, hardening, fuzz, race, leak, mutation,
  compatibility, and comparative benchmark coverage.
- Adversarial terminal-path, weighted-starvation, concurrent partition
  replacement, Kubernetes lifecycle-model, and cross-package resilience
  composition campaigns, plus wait-latency, fairness, cancellation, observer,
  partition, throughput, and maintained-implementation benchmarks.
