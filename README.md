# SpotSync

> A high-concurrency reservation engine for finite parking and EV charging resources — built to **never oversell** a zone under contention.

[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)

- **Live API:** _pending deployment_
- **API base path:** `/api/v1`

---

## Overview

SpotSync is a Go backend (Echo + GORM + PostgreSQL) that implements the Apollo B6A6 graded contract: nine REST endpoints, JWT auth, role-based access, and concurrency-safe reservations with dynamic availability.

See SpotSync_Flagship_Blueprint.md for architecture strategy and SpotSync_Execution_Plan.md for the phased build Plan.

---

## Tech stack

| Layer | Choice |
| --- | --- |
| Language | Go 1.22+ |
| HTTP | Echo v4 |
| ORM | GORM (PostgreSQL) |
| Auth | JWT + bcrypt |
| Validation | go-playground/validator |
| Migrations | golang-migrate |

---

## Project structure

```
cmd/api/          API entrypoint
cmd/worker/       Release worker (Phase 1+)
cmd/seed/         Admin bootstrap seed
internal/         Clean Architecture layers
migrations/       Versioned SQL migrations
deploy/           Docker, Compose, K8s manifests
test/             Contract, integration, and load tests
docs/             Architecture docs and ADRs
```

---

## Getting started

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- golang-migrate CLI
- Optional: air (hot reload), golangci-lint, lefthook, Docker

### Setup

```bash
cp .env.example .env
make migrate-up
make run
```

Common tasks: `make help`

---

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| PORT | no | HTTP port (default 8080) |
| DATABASE_URL | yes | PostgreSQL connection string |
| JWT_SECRET | yes | JWT signing secret |
| JWT_EXPIRY | no | Token lifetime (default 24h) |
| BCRYPT_COST | no | bcrypt cost (default 12) |
| ALLOW_SELF_ADMIN_REGISTRATION | no | Honor admin role on register (default true) |
| CORS_ALLOWED_ORIGINS | no | Comma-separated allowed origins |
| LOG_LEVEL | no | Log verbosity (info) |

---

## API endpoints

| Method | Path | Access |
| --- | --- | --- |
| POST | /api/v1/auth/register | Public |
| POST | /api/v1/auth/login | Public |
| POST | /api/v1/zones | Admin |
| GET | /api/v1/zones | Public |
| GET | /api/v1/zones/:id | Public |
| POST | /api/v1/reservations | Authenticated |
| GET | /api/v1/reservations/my-reservations | Authenticated |
| DELETE | /api/v1/reservations/:id | Authenticated |
| GET | /api/v1/reservations | Admin |

Full request/response examples will be documented in Phase 0 step 12.

---

## Development

```bash
make fmt
make vet
make lint
make test
make test-race
make dev
```

Install git hooks: lefthook install

---

## License

MIT — see LICENSE.
