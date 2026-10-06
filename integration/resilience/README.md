# Bulkhead resilience composition

This non-releasable integration module proves the application-owned resilience
ordering through the public bulkhead, retry, and circuit-breaker contracts.
Local bulkhead admission failure remains a permanent retry outcome and occurs
before circuit-breaker admission, so it neither amplifies attempts nor records
a downstream failure. The in-process operation uses Retry v2.1's strict outcome
contract and reports `OutcomeKnown` because bulkhead rejection conclusively
proves that no downstream operation was dispatched.

The attached Resilience v2 scope admits one additional retry with original/retry
lineage, then refuses further dispatch. Completed attempts release their permits;
bulkhead capacity remains reusable.
Resilience v1 remains an indirect dependency of Retry's intentionally retained
legacy-budget support; these tests do not attach a legacy scope.

Run from this module with `GOWORK=off`:

```sh
go test ./...
```
