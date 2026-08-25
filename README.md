# Sky KV

**Status: engineering beta.** Sky KV is a small dependency-free Go HTTP service for bounded, process-local key/value state.

## Verified product surface

- `POST /v1/kv` creates or updates a key with an optional `expected_version` optimistic-concurrency check.
- `GET /v1/kv?key=...` reads the value, version and update timestamp.
- `DELETE /v1/kv?key=...&version=...` supports optional compare-and-delete semantics.
- `GET /healthz`, `GET /readyz`, and `GET /metrics` expose operational state.
- Keys are trimmed and bounded to 256 characters; values are bounded to 64 KiB; request bodies are capped.
- `MAX_KEYS` bounds in-memory cardinality (default 10,000).
- `sync.RWMutex` protects concurrent access and the race detector is part of CI.
- HTTP server timeouts, strict JSON decoding, Go vet/tests, `govulncheck`, container build, and non-root image verification are CI gates.

## Run locally

```bash
go test -race ./...
go run .
```

The service listens on `:8080` by default. `PORT` and `MAX_KEYS` are configurable.

Example:

```bash
curl -sS -X POST http://127.0.0.1:8080/v1/kv \
  -H 'content-type: application/json' \
  -d '{"key":"profile:42","value":"active"}'

curl -sS 'http://127.0.0.1:8080/v1/kv?key=profile:42'
```

## Container

```bash
docker build -t sky-kv .
docker run --rm -p 8080:8080 sky-kv
```

The runtime image is distroless and executes as a non-root user.

## SKYCOIN4444 integration

Sky KV can back bounded ephemeral state such as local caches, idempotency hints, or development adapters through its HTTP contract. Durable ecosystem data should remain behind a persistent database service rather than depending on this process-local store.

## Explicit boundaries

Sky KV does **not** claim persistence, replication, distributed consensus, cross-process locking, Redis compatibility, authentication/RBAC, encryption at rest, HA, multi-region consistency, or production deployment. Process restart loses all data. Those capabilities require separate implementations and evidence.

See `SECURITY.md` for the current security boundary.
