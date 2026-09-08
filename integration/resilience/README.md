# Bulkhead resilience composition

This non-releasable integration module proves the application-owned resilience
ordering through the public bulkhead, retry, and circuit-breaker contracts.
Local bulkhead admission failure remains a permanent retry outcome and occurs
before circuit-breaker admission, so it neither amplifies attempts nor records
a downstream failure. The in-process operation uses Retry v1.1's strict outcome
contract and reports `OutcomeKnown` because bulkhead rejection conclusively
proves that no downstream operation was dispatched.

Run from the repository workspace with:

```sh
go test ./integration/resilience/...
```
