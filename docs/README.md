# Documentation

## Getting started

- [Install and quick start](../README.md#install)
- [Package and internal harness map](../README.md#package-map)
- [Executable examples](../example_test.go)
- [Package API](https://pkg.go.dev/github.com/faustbrian/go-bulkhead)

## Guides

| Guide | Purpose |
| --- | --- |
| [API](api.md) | Construction, admission, permits, execution, errors, observations, and drain |
| [Architecture](architecture.md) | Ownership, synchronization, fairness, bounds, and dependency boundary |
| [Composition](composition.md) | Retry, breaker, rate, adaptive limit, hedge, cache, timeout, and fallback order |
| [Kubernetes](kubernetes.md) | Per-pod sizing, HPA, rollout, readiness, and termination |
| [Operations](operations.md) | Metrics, alerts, troubleshooting, incidents, and drain runbook |
| [Migration](migration.md) | Channels, `x/sync/semaphore`, worker pools, and unbounded registries |
| [Security](security.md) | Identity bounds, cardinality, denial-of-service, and secret handling |
| [Performance](performance.md) | Equivalent-behavior benchmarks and semantic differences |
| [Assurance](assurance.md) | Requirement mapping, concurrency invariants, and current evidence |
| [FAQ](faq.md) | Selection and behavior questions |

## Support and maintenance

- [Saturation troubleshooting](operations.md#incident-runbook)
- [Support](../SUPPORT.md)
- [Security policy and reporting guidance](../SECURITY.md)
- [Compatibility policy](../COMPATIBILITY.md)
- [Contribution guide](../CONTRIBUTING.md)
- [Release history](../CHANGELOG.md)
- [License](../LICENSE)
