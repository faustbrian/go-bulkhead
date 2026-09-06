# Security policy

## Supported versions

The latest stable v1 release receives security fixes. Older releases and the
`main` branch are unsupported; upgrade before reporting unless the issue is a
regression under active development.

| Version | Supported |
| --- | --- |
| Latest stable v1 release | Yes |
| Older releases | No |
| `main` | No |

## Reporting a vulnerability

Do not disclose a suspected vulnerability in a public issue. Use the
[private vulnerability reporting form](https://github.com/faustbrian/go-bulkhead/security/advisories/new).

Do not include credentials, production payloads, customer identifiers, or
arbitrary resource labels in a public report or initial contact request.

Bulkheads are process-local availability controls, not authorization or
distributed coordination. Applications must bound resource identity, partition
count, queue size, wait, execution context, fan-out, retries, hedges, and
maximum replica capacity.
