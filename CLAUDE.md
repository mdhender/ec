# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

EC (Epimethean Challenge) is a multiplayer strategy game server. It is a monorepo containing a Go API server, a Go CLI, a React frontend, and Hugo-based documentation. The API server and CLI share one SQLite database.

## Build Commands

```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/api/
cd backend && go build ./cmd/cli/
```

## Architecture: SOUSA

This project enforces a strict inward dependency rule. The canonical reference is `docs/SOUSA.md`.

```
domain ← app ← infra / delivery ← runtime
```

| Layer | Path | Responsibility |
|---|---|---|
| `domain` | `backend/internal/domain/` | Pure game rules, entities, deterministic logic — no framework imports |
| `cerr` | `backend/internal/cerr/` | Canonical sentinel errors with business meaning, not transport formatting |
| `app` | `backend/internal/app/` | Use cases, port interfaces, orchestration, transaction intent |
| `infra` | `backend/internal/infra/` | Concrete adapters: SQLite, filestore, auth — implements `app` interfaces |
| `delivery` | `backend/internal/delivery/` | HTTP (Echo) handlers and CLI commands — thin adapters only |
| `runtime` | `backend/internal/runtime/` | Wires concrete types together; owns startup and operational mode |

Go `main` packages live in `backend/cmd/api/` and `backend/cmd/cli/`. The `apps/` directory contains non-Go artifacts only (`apps/web/` for the React frontend and `apps/site/` for the Hugo static site).

**Hard rules (never violate):**
- `domain` must not import `app`, `infra`, `delivery`, or `runtime`.
- `app` must not import Echo, SQLite packages, CLI frameworks, or filesystem adapters.
- `infra` implements ports defined by `app`; `app` never depends on infra concrete types.
- `delivery` must not import `infra`. They are peers — both depend inward on `app`, never on each other.
- Both `delivery/http` and `delivery/cli` are peers and must both remain thin — no game logic, no embedded SQL.
- The CLI is not privileged; it calls the same `app` layer as the HTTP API.
- `runtime` is the only layer that instantiates and injects concrete implementations.
- `cmd/` packages must not import `infra` directly; they pass raw config to `runtime`, which owns wiring.

## Operational Model

The API server and CLI share one SQLite database. Turn processing is a batch workflow: the API server is stopped (or put in maintenance mode) before the CLI runs turn-processing commands, giving the CLI exclusive write access. This is an intentional design constraint, not a defect. Coordination logic belongs in `runtime` and deployment scripts — not in `domain` or HTTP handlers.

## Testing Strategy

- **Domain tests**: fast, no DB/filesystem, cover invariants and edge cases.
- **App tests**: use-case orchestration via interfaces; may use test doubles.
- **Infra tests**: real SQLite databases and temp directories; cover SQL correctness and migrations.
- **Delivery tests**: request parsing, error mapping, JSON contracts, exit codes.
- **Runtime tests**: use sparingly; only startup wiring and critical config failures.

## Frontend & Docs

- **`apps/web/`** — React + Vite + TailwindCSS. Client-side UX only; backend is authoritative for all game rules and authorization.
- **`apps/site/`** — Hugo + Hextra static site. Must reflect actual system behavior; examples should be checked against real API behavior.
