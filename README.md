# Go Key-Value Store

Small Go service providing an in-memory key/value store over HTTP.

## Implemented

- Thread-safe in-memory storage using `sync.RWMutex`
- `POST /set` JSON endpoint
- `GET /get?key=...` lookup endpoint
- `GET /health` health endpoint
- HTTP read/write timeouts
- Go test/CI scaffolding already present in the repository

## Limitations

This is **not a durable production database**. Data is lost when the process stops. It does not currently provide persistence, replication, authentication, authorization, encryption at rest, or a production-grade distributed consistency model.

For SKYCOIN4444 it is treated as a reusable cache/prototype service capability, not the canonical production database.

## Ecosystem role

Potential canonical boundary: **Supporting Services / Cache**. The production persistence boundary belongs in the canonical Database layer.

## Validation

The repository contains Go tests and CI configuration. Passing status must be established from actual workflow/test evidence; this README does not claim that checks currently pass.

## License

See the repository license and existing source files for applicable terms.
