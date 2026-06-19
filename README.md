# Go-Key-Value-Store

![CI](https://github.com/skylerblue333/Go-Key-Value-Store/workflows/CI/badge.svg)

High-performance, thread-safe, in-memory key-value store written in Go.

## Features
- `sync.RWMutex` for highly concurrent reads
- REST API (`/get`, `/set`)
- Dockerized multi-stage build
- 100% Test Coverage

## Quick Start
```bash
go test ./...
go run main.go
```
