# CampusCare AI - Phase 0

Production-grade foundation for the CampusCare AI platform.

## Architecture

This project follows a modular, clean architecture pattern:
- `cmd/api`: Application entrypoint
- `internal/config`: Configuration management
- `internal/domain`: Core domain models (User, Role)
- `internal/database`: PostgreSQL connection management
- `internal/delivery/http`: API routing, handlers, middleware
- `internal/server`: HTTP Server lifecycle
- `pkg/`: Shared utilities

## Prerequisites
- Go 1.21+
- PostgreSQL

## Setup
1. Copy `.env.example` to `.env` and configure DB settings.
2. Run `go mod tidy`.
3. Start the application: `make run`.
