# SOUSA Guidelines For EC (Epimethean Challenge)

SOUSA is a strict placement and import discipline. Dependencies flow inward only:

```
domain ← app ← infra / delivery ← runtime
```

Outer layers may depend on inner layers. Inner layers must never depend on outer layers.

---

## Repository Layout

```
repo/
  backend/
    go.mod
    cmd/
      api/          ← main package for the HTTP server
      cli/          ← main package for the CLI
    internal/
      domain/
      cerr/
      app/
      infra/
        sqlite/
        filestore/
        auth/
      delivery/
        http/
        cli/
      runtime/
        server/
        cli/

  apps/
    web/            ← React + Vite + TailwindCSS
      src/
        lib/
        components/
        pages/
    site/           ← Hugo + Hextra static site
      content/
      assets/
      layouts/

  ops/
    caddy/
    systemd/
    deploy/

  scripts/
```

All Go source lives under `backend/`. The `apps/` directory contains non-Go artifacts only.

---

## Layer Rules

### domain — `backend/internal/domain/`

Contains game entities, invariants, turn-state transformations, and validation intrinsic to the game itself. Must be deterministic given the same inputs.

**Must not import:** `app`, `infra`, `delivery`, `runtime`, Echo, SQLite packages, filesystem adapters, CLI frameworks, or HTTP types.

### cerr — `backend/internal/cerr/`

Defines canonical sentinel errors shared across the backend. Errors express business/application meaning, not transport formatting.

Examples: `ErrGameNotFound`, `ErrUnauthorized`, `ErrInvalidOrders`, `ErrTurnAlreadyProcessed`.

**Must not import:** `app`, `infra`, `delivery`, or `runtime`.

### app — `backend/internal/app/`

Contains use cases, port interfaces, and orchestration. Defines transaction intent. Coordinates domain logic through interfaces declared here; never through concrete types.

**May import:** `domain`, `cerr`, and standard library.

**Must not import:** Echo, SQLite packages, filesystem adapters, CLI frameworks, or HTTP request/response types.

### infra — `backend/internal/infra/`

Concrete adapters that implement interfaces declared in `app`.

- `infra/sqlite` — database open/close, pragmas, migrations, repositories, transactions, SQL.
- `infra/filestore` — upload storage, generated reports, exports, path conventions.
- `infra/auth` — token signing/verification, magic link persistence, session/cookie adapters.

**Must not expose** concrete types to `app` or `domain`. Inner layers depend only on the interfaces, not the implementations.

### delivery — `backend/internal/delivery/`

Translates external input/output into application calls. `delivery/http` and `delivery/cli` are peers; neither is privileged.

**Permitted actions:** parse input, validate transport-level shape, call application use cases, map errors, format output.

**Must not:** implement game rules, embed SQL, make filesystem decisions, or duplicate orchestration that belongs in `app`.

- `delivery/http` — Echo route definitions, request parsing, error-to-HTTP mapping, JSON serialization. DTOs are delivery contracts, not domain models.
- `delivery/cli` — command/flag parsing, calling use cases, printing output, returning exit codes. The CLI calls the same `app` layer as the HTTP server.

### runtime — `backend/internal/runtime/`

Wires concrete implementations together into runnable processes. The only layer that instantiates concrete adapters and injects them.

- `runtime/server` — loads config, opens DB, creates repos/services, wires Echo routes, starts HTTP server.
- `runtime/cli` — loads config, opens DB, creates repos/services, wires CLI commands, executes selected command.

**Must not** be imported by any inner layer.

---

## Shared SQLite Operational Model

The API server and CLI share one SQLite database file. Turn processing is a batch workflow: the API server is stopped or placed in maintenance mode before the CLI runs turn-processing commands. The CLI holds exclusive write access during turn execution. This is an intentional constraint, not a defect.

Enforcement of this rule belongs in `runtime`, deployment scripts, and operator documentation — not in `domain`, `app`, or delivery handlers.

Transaction intent is defined in `app`. SQLite transaction mechanics are implemented in `infra/sqlite`. Write concurrency must be treated conservatively; batch workflows must assume exclusive write ownership.

---

## Frontend Rules

`apps/web/` — React + Vite + TailwindCSS.


The frontend may: validate form shape, provide client-side UX logic, render server state.

The frontend must not be the authoritative source for: turn resolution, authorization policy, game invariants, or hidden operational rules. The backend is authoritative.

---

## Documentation Rules

`apps/site/` — Hugo + Hextra.

Examples must be checked against real API behavior. Terminology must match domain and application language. Operator docs must match runtime and deployment behavior.

---

## Testing by Layer

| Layer | Scope | Infrastructure |
|---|---|---|
| `domain` | invariants, rules, edge cases | none — no DB or filesystem |
| `app` | use-case orchestration, transaction behavior | test doubles or narrow integration harness |
| `infra` | SQL correctness, repository behavior, migrations | real SQLite DB, temp directories |
| `delivery/http` | request parsing, error mapping, JSON contracts | — |
| `delivery/cli` | flag parsing, command wiring, exit codes | — |
| `runtime` | startup wiring, critical config failures | use sparingly |

---

## Hard Rules

1. `domain` must not import `app`, `infra`, `delivery`, or `runtime`.
2. `app` must not import Echo, SQLite concrete packages, CLI frameworks, or filesystem adapters.
3. `infra` implements ports defined by `app`; `app` never depends on infra concrete types.
4. `delivery/http` and `delivery/cli` are peers; both must remain thin.
5. The CLI must not bypass the application layer for core game behavior.
6. The API must not implement game rules independently of the core.
7. `runtime` owns wiring and operational mode selection; no inner layer imports `runtime`.
8. The shared SQLite operational model must be explicit in deployment scripts and operator docs.
9. Frontend code must not be the authoritative source of business rules or game logic.
10. Documentation must reflect actual system behavior.
