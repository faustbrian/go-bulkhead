# Changelog

All notable changes to this module are documented here.

## Unreleased

### Documentation

- Keep the README focused on adoption and move detailed assurance guidance to
  the package documentation index.

## 1.0.0 - 2026-08-25

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
