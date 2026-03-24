# SOUSA for Humans

This document explains the architecture of EC (Epimethean Challenge) for people, not coding agents. The authoritative rules document is `docs/SOUSA.md`.

---

## For Developers

### The core idea

SOUSA is a layering discipline. Code is organized into layers, and dependencies only flow inward — outer layers know about inner layers, never the reverse.

```
domain ← app ← infra / delivery ← runtime
```

Think of it as an onion. The game rules are at the center. Everything else wraps around them.

### The layers and what goes where

**`domain`** is the game engine. Entities, rules, turn resolution, invariants. Pure Go — no HTTP, no SQL, no framework types. If you're writing logic that answers "what is valid in this game?", it goes here.

**`cerr`** is a small shared package of named errors: `ErrGameNotFound`, `ErrUnauthorized`, `ErrTurnAlreadyProcessed`, and so on. Errors here carry business meaning. They don't know whether they'll be turned into a 404 or a CLI exit code — that's someone else's job.

**`app`** contains use cases: submit orders, run a turn, authenticate, publish results. It calls `domain` for game logic and talks to infrastructure through interfaces it defines. It never imports a concrete database driver or an HTTP library. If you're writing "what does the system do?", it goes here.

**`infra`** is where the concrete plumbing lives. Three sub-packages:
- `infra/sqlite` — all SQL, migrations, repositories, transactions.
- `infra/filestore` — uploads, generated reports, exports, file layout.
- `infra/auth` — token signing, magic link handling, session details.

Infra implements the interfaces that `app` defines. `app` never depends on infra directly — it only knows about the interface, not the implementation behind it.

**`delivery`** is where the outside world enters. Two peers:
- `delivery/http` — Echo handlers. Parse the request, call a use case, map the result to JSON. Thin.
- `delivery/cli` — Commands. Parse flags, call a use case, print output. Also thin.

Neither delivery layer implements game logic. The CLI is not a privileged back door — it calls the same `app` layer that the HTTP server does.

**`runtime`** wires everything together. It's the only layer that knows about concrete implementations and injects them where needed. `runtime/server` starts the HTTP server; `runtime/cli` runs CLI commands.

### Where to put new code

| You're writing... | It goes in... |
|---|---|
| A new game rule or invariant | `domain` |
| A new user action or workflow | `app` (new use case + interface if needed) |
| A new SQL query or repository | `infra/sqlite` |
| A new HTTP endpoint | `delivery/http` |
| A new CLI command | `delivery/cli` |
| A new auth mechanism | `infra/auth` |
| Startup/config wiring | `runtime/server` or `runtime/cli` |

### The Go module layout

All Go source lives under `backend/`. The `go.mod` is at `backend/go.mod`. The entry points are `backend/cmd/api/` (HTTP server) and `backend/cmd/cli/` (CLI). Non-Go apps (`apps/web/`, `apps/site/`) live outside and are built separately.

### The SQLite constraint

Both the HTTP server and the CLI share one SQLite database file. SQLite handles concurrent reads fine, but only one writer at a time. Turn processing is a batch workflow that writes heavily — when it runs, the HTTP server is stopped (or put in maintenance mode) first. This is a known, intentional constraint. Don't try to work around it in code; it's handled at the deployment level.

---

## For Testers

### How testing maps to the layers

Each layer has a distinct testing character. Understanding this helps you know what kind of test to write and what setup it needs.

**Domain tests** are the fastest and simplest. They test pure logic: given this game state and these orders, does the rule produce the right result? No database, no filesystem, no HTTP. Just function calls. These should be numerous and exhaustive — they're cheap to run and protect the most important behavior.

**App tests** test use-case orchestration. Does submitting orders in the wrong phase return the right error? Does a turn get marked as processed after running? These tests work against interfaces, so infrastructure can be replaced with a test double or an in-memory stand-in. The goal is to verify the workflow, not the SQL.

**Infra tests** test the concrete adapters. Does the SQLite repository return the right rows? Do migrations apply cleanly? Does the filestore write to the expected path? These tests use real SQLite databases in temporary directories. They're slower but necessary — they verify the actual data layer.

**Delivery tests** test the transport boundary. For HTTP: does a malformed request return a 400? Does an unauthorized request return a 401? Does the JSON response match the documented contract? For CLI: does the command parse its flags correctly and exit with the right code? These tests don't need real game logic — they test the translation layer.

**Runtime tests** are minimal. They check that the process starts up correctly with valid config and fails fast with invalid config. Don't over-invest here.

### What testers should watch for

The CLI and HTTP server share the same application core. A behavior confirmed in one should be consistent in the other — any divergence is a bug. If a rule is enforced by the HTTP API but not the CLI (or vice versa), the architecture has been violated somewhere.

The frontend is not authoritative. Authorization and game rule validation in the browser are UX conveniences, not security boundaries. Tests for correctness belong on the backend.

During turn processing the API is offline. Any tests that simulate the operational sequence (stop API → run CLI turn → restart API) need to account for this.

---

## For Project Managers

### What the architecture is trying to achieve

EC has two ways to interact with the game: a web API (used by players through a browser) and a command-line tool (used by operators to run turns, import data, and manage the game). Both need to behave consistently — the game rules must be the same regardless of how you're accessing the system.

SOUSA enforces this by putting the game rules in one place that both interfaces must call. The CLI cannot quietly skip a rule that the API enforces, and the API cannot add a rule the CLI doesn't know about. There is one implementation of game behavior, full stop.

### The artifacts this project ships

| Artifact | What it is |
|---|---|
| HTTP API server | The backend service players connect to |
| CLI | Operator tooling for turn processing, imports, admin tasks |
| `apps/web/` | The player-facing web interface (React) |
| `apps/site/` | Public documentation website (Hugo) |

The API server and CLI are compiled from the same Go codebase and share a database. The web interface and documentation site are static builds deployed separately.

### The one operational constraint worth knowing

Turn processing (advancing the game by one turn) is a batch job run by the CLI. While it runs, the API server is taken offline briefly. This is a deliberate tradeoff: it keeps the system simple and the game state consistent. The downtime window is predictable and operator-controlled, not an outage. This constraint is managed at the deployment level and does not require changes to the game logic.

### Why this approach

A common failure mode in game servers is that rules drift — the server enforces one version of a rule, the client assumes another, and admin tools do something different again. SOUSA prevents this structurally, not through discipline or documentation alone. The architecture makes it difficult to put rules in the wrong place.

The tradeoff is some upfront strictness about where code goes. The benefit is that the game engine stays testable and consistent as the project grows.
