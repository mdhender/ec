# SOUSA Guidelines For EC (Epimethean Challenge)

SOUSA is a strict architecture policy inspired by Onion/Clean Architecture. In EC, it means game rules and business logic stay in pure, testable code, while delivery and I/O concerns (HTTP, SQLite, JWT, filesystem, React UI, runtime wiring) stay in outer adapters.

## Purpose

This document defines the architectural rules for this project using a SOUSA-style layout.

For coding agents, SOUSA is a placement and import discipline:

- put behavior in the correct layer
- keep dependencies flowing one way
- do not bypass boundaries for convenience

This document defines how coding agents should apply SOUSA in this repository.

The goal is to keep game rules, application behavior, delivery mechanisms, and operational concerns clearly separated so that:

- the game engine remains deterministic and testable,
- the API server and CLI share the same core behavior,
- the frontend stays a thin client over documented contracts,
- SQLite and filesystem details remain replaceable infrastructure,
- deployment and turn-processing workflows stay simple and explicit.

This is a green-field architecture intended for a project with:

- static documentation built with Hugo + Hextra,
- a static frontend built with React + Vite + TailwindCSS,
- a Go API server using Echo, SQLite3, and a REST-ish JSON API,
- a Go CLI that implements the game and administrative workflows,
- a shared SQLite database used by both the API server and the CLI.

---

## Architectural Principles

### 1. Dependencies point inward

Code in outer layers may depend on inner layers.
Code in inner layers must not depend on outer layers.

Allowed direction:

```text
domain <- app <- infra / delivery <- runtime
````

Where:

* `domain` contains pure business rules and entities,
* `app` contains use cases, ports, and orchestration,
* `infra` contains adapters for SQLite, filesystem, auth, and similar details,
* `delivery` contains HTTP and CLI entry points,
* `runtime` wires concrete implementations together.

### 2. The game rules live in one place

There must be exactly one authoritative implementation of game behavior.

The CLI is not allowed to implement “special real logic” that bypasses the application layer.
The API server is not allowed to implement its own rules independently of the CLI.

Both must call the same application and domain code.

### 3. Delivery layers stay thin

HTTP handlers and CLI commands are delivery code.
They may:

* parse input,
* validate transport-level shape,
* call application use cases,
* map errors,
* format output.

They must not:

* implement game rules,
* embed SQL,
* make ad hoc filesystem decisions,
* duplicate orchestration that belongs in `app`.

### 4. Infrastructure stays outside the core

SQLite, report storage, auth token storage, file layout, and similar concerns are infrastructure details.

The application layer defines the interfaces it needs.
Infrastructure implements those interfaces.

### 5. Runtime owns process wiring and operational mode

The fact that the API server and CLI share a SQLite database is an operational concern, not a domain concern.

Rules such as:

* API runs during interactive mode,
* API is stopped or placed in maintenance mode during turn processing,
* CLI gets exclusive write access during turn execution,

belong in runtime and deployment tooling, not in domain logic.

---

## Repository Strategy

This project uses a **monorepo**.

Rationale:

* the API, CLI, frontend, and docs are part of one product,
* the API and CLI share a database and core behavior,
* the frontend depends on the API contract,
* the docs must track the behavior of the system,
* coordinated changes should be atomic.

This repository may produce multiple independent artifacts, but they are versioned together.

---

## Recommended Repository Layout

```text
repo/
  apps/
    api/
      cmd/server/
    cli/
      cmd/cli/
    web/
      src/
        lib/
        components/
        pages/
    docs/
      content/
      assets/
      layouts/

  backend/
    go.mod
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

  ops/
    caddy/
    systemd/
    deploy/

  scripts/
```

This layout is illustrative.
Names may change, but the dependency rules may not.

---

## Layer Definitions

## Domain

### Responsibility

The `domain` layer defines the game’s core concepts and rules.

Examples:

* entities and value objects,
* invariants,
* deterministic rule evaluation,
* turn-state transformations,
* validation that is intrinsic to the game itself.

### Rules

The domain layer must:

* be pure Go logic,
* be deterministic when given the same inputs,
* avoid knowledge of HTTP, CLI flags, SQL, JSON transport, filesystem paths, and framework types.

The domain layer must not import:

* Echo,
* SQLite packages,
* filesystem adapters,
* frontend contracts,
* runtime/config packages.

### Examples of things that belong here

* game entities,
* turn resolution rules,
* order validation rules that are intrinsic to gameplay,
* canonical identifiers and value semantics,
* calculations and transformations.

### Examples of things that do not belong here

* SQL queries,
* HTTP status codes,
* CLI output formatting,
* JSON response shapes,
* file path conventions,
* deployment flags.

---

## cerr

### Responsibility

The `cerr` layer defines canonical errors used across the backend.

This may include:

* domain-relevant sentinel errors,
* application-level classification errors,
* mapping helpers for delivery layers.

The purpose is to keep error meaning stable even when delivery mechanisms differ.

### Rules

Errors here should describe business/application meaning, not transport formatting.

For example:

* `ErrGameNotFound`
* `ErrUnauthorized`
* `ErrInvalidOrders`
* `ErrTurnAlreadyProcessed`

They should not encode HTTP response formatting or CLI presentation.

---

## App

### Responsibility

The `app` layer contains application use cases and orchestration.

It coordinates domain logic and infrastructure through interfaces defined here.

Examples:

* submit orders,
* list reports,
* authenticate by magic link,
* import uploaded turn report,
* run turn,
* publish turn outputs,
* mark game state transitions.

### Rules

The application layer may depend on:

* `domain`,
* `cerr`,
* standard library packages appropriate for orchestration.

The application layer must not depend on:

* Echo,
* SQLite concrete implementations,
* concrete filesystem layout,
* CLI frameworks,
* HTTP request/response types.

### Ports

Interfaces required by the application layer are defined here.

Examples:

* game repositories,
* unit-of-work or transaction boundaries,
* report storage,
* auth token stores,
* clock or randomness abstractions where needed,
* audit/log sinks if necessary.

### Transactions

Application use cases are responsible for defining transactional boundaries.

The exact SQLite transaction behavior is implemented in infrastructure, but transaction intent belongs here.

---

## Infra

### Responsibility

The `infra` layer contains concrete adapters.

These adapters implement interfaces defined by the application layer.

Likely subpackages:

* `infra/sqlite`
* `infra/filestore`
* `infra/auth`

### SQLite adapter

`infra/sqlite` owns:

* opening the database,
* pragmas and connection configuration,
* migrations integration,
* repository implementations,
* transaction implementations,
* SQL statements and query helpers.

SQLite details must not leak into `app` or `domain`.

### Filestore adapter

`infra/filestore` owns:

* upload storage,
* generated reports,
* export files,
* path conventions,
* filesystem safety and organization.

### Auth adapter

`infra/auth` owns:

* token signing/verification details,
* magic link persistence or validation,
* password hashing if ever added,
* cookie/session/token adapter details.

### Rules

Infrastructure may depend inward on `app` and `domain` contracts.
Inner layers must never depend on infrastructure concrete types.

---

## Delivery

Delivery code translates external input/output into application calls.

This project has two first-class delivery mechanisms:

* HTTP delivery
* CLI delivery

They are peers.

```text
delivery/
  http/
  cli/
```

## Delivery / HTTP

### Responsibility

The HTTP delivery layer contains Echo handlers and transport mapping.

It is responsible for:

* route definitions,
* request parsing,
* transport-level validation,
* calling use cases,
* mapping errors to HTTP responses,
* JSON serialization.

### Rules

HTTP handlers must remain thin.

They must not:

* contain embedded SQL,
* decide filesystem layout,
* duplicate game rules,
* perform turn-processing orchestration that belongs in `app`.

### Notes

REST-ish JSON DTOs are delivery/API contract concerns.
They may map to domain/application structures, but they are not themselves domain models.

## Delivery / CLI

### Responsibility

The CLI delivery layer contains commands for:

* administrative workflows,
* turn processing,
* imports/exports,
* maintenance operations,
* operator-facing tooling.

### Rules

CLI commands are delivery code, not privileged shortcuts.

They may:

* parse flags and arguments,
* call application use cases,
* print status and progress,
* return appropriate exit codes.

They must not:

* bypass application rules,
* embed SQL,
* contain separate game logic,
* perform ad hoc writes that ignore repositories and transactions.

### Special note

The CLI is a first-class delivery mechanism because it implements operational and batch workflows.
That does not give it permission to bypass SOUSA boundaries.

---

## Runtime

### Responsibility

The `runtime` layer wires the system together into executable processes.

Likely subpackages:

* `runtime/server`
* `runtime/cli`

Runtime is where concrete adapters are instantiated and injected.

### Runtime / server

Responsible for:

* loading config,
* opening the SQLite database,
* creating repositories and services,
* wiring Echo routes,
* starting the HTTP server.

### Runtime / cli

Responsible for:

* loading config,
* opening the SQLite database,
* creating repositories and services,
* wiring CLI commands,
* executing the selected command.

### Rules

Runtime may know about concrete implementations.
Inner layers must not know about runtime.

---

## Frontend Architecture

The frontend is a static web application built with React + Vite + TailwindCSS.

Recommended structure:

```text
apps/web/
  src/
    lib/
    components/
    pages/
```

### `src/lib`

Contains:

* API client code,
* auth helpers,
* shared UI-side utilities,
* query/fetch logic,
* DTO handling specific to the frontend.

### `src/components`

Contains reusable UI components.

### `src/pages`

Contains page-level composition.

### Rules

The frontend must not become a second business-rules engine.

It may:

* validate form shape,
* provide client-side UX logic,
* render server state.

It must not become the source of truth for:

* turn resolution,
* authorization policy,
* game invariants,
* hidden operational rules.

The backend remains authoritative.

---

## Documentation Architecture

Documentation is built statically with Hugo + Hextra.

Recommended structure:

```text
apps/docs/
  content/
  assets/
  layouts/
```

### Role of documentation

Documentation is not runtime code, but it is product-critical.

It should cover:

* user documentation,
* operator documentation,
* API documentation,
* architecture and development guidance,
* game rule references,
* implementation-facing manuals where needed.

### Rules

Docs must track the actual system.

Where practical:

* examples should be checked against real API behavior,
* terminology should match domain and application language,
* operator docs should match runtime and deployment behavior,
* rule docs should identify canonical sources of truth.

---

## Shared Database Model

The API server and CLI share the same SQLite database.

This is an intentional design choice.

### Architectural stance

* SQLite is the current system of record.
* The API and CLI are separate processes.
* Turn processing is a batch workflow with significant write activity.
* Exclusive write ownership during turn execution is acceptable and expected.

### Operational rule

The deployment/runtime workflow may stop the API server, or place it into maintenance mode, before running CLI turn-processing commands.

This is not considered an architectural defect.
It is a valid operational constraint for this system.

### Placement of concern

This rule belongs in:

* `runtime`,
* deployment scripts,
* operator documentation.

It does not belong in:

* `domain`,
* HTTP handlers,
* CLI command business logic.

---

## Transactions and Concurrency

Because SQLite is shared by multiple processes, write concurrency must be treated conservatively.

### Rules

* application use cases define transaction intent,
* infrastructure implements SQLite transaction mechanics,
* batch/turn-processing workflows must assume exclusive write ownership,
* interactive API workflows should remain short-lived and transactional.

If the architecture changes later to use another datastore, the application and domain layers should remain stable.

---

## Contract Rules

## Backend contract

The API contract is defined at the delivery boundary.
It should be stable, versionable, and documented.

DTOs should not be treated as domain entities.

## Frontend contract

Frontend code depends on documented API behavior, not on backend internals.

## CLI contract

CLI commands are operator-facing contracts.
They should be stable enough for scripts, but they do not define domain truth.

---

## Testing Strategy by Layer

## Domain tests

Focus on:

* invariants,
* deterministic rules,
* edge cases,
* canonical examples.

These should be fast and require no database or filesystem.

## App tests

Focus on:

* use-case orchestration,
* transactional behavior,
* repository interaction through interfaces,
* error semantics.

These may use test doubles or narrow integration harnesses.

## Infra tests

Focus on:

* repository correctness,
* SQL behavior,
* transaction behavior,
* migration compatibility,
* filesystem adapter correctness.

These may use real SQLite databases and temporary directories.

## Delivery tests

### HTTP

Focus on:

* request parsing,
* error mapping,
* JSON contract behavior,
* route-level integration.

### CLI

Focus on:

* flag parsing,
* command wiring,
* exit codes,
* output contract where important.

## Runtime tests

Use sparingly.
Focus on startup wiring and critical configuration failures.

---

## Non-Goals

This architecture does not aim to:

* make every package reusable outside the project,
* optimize for microservice decomposition,
* eliminate all operational coordination,
* hide the existence of SQLite constraints,
* move game logic into the frontend.

It aims to produce a clear, disciplined, maintainable system.

---

## Hard Rules

1. `domain` must not import `app`, `infra`, `delivery`, or `runtime`.
2. `app` must not import Echo, SQLite concrete packages, CLI frameworks, or filesystem adapters.
3. `infra` implements ports defined by `app`; `app` does not depend on infra concrete types.
4. `delivery/http` and `delivery/cli` are peers and must both remain thin.
5. The CLI must not bypass the application layer for core game behavior.
6. The API must not implement game rules independently of the core.
7. Runtime owns wiring and operational mode selection.
8. Documentation must reflect the actual behavior of the system.
9. Frontend code must not become the authoritative source of business rules.
10. The shared SQLite operational model must be explicit in scripts and docs.

---

## Decision Record

At the time this document was written, the project intentionally chooses:

* a monorepo,
* one shared backend/core architecture for API and CLI,
* a static frontend,
* static documentation,
* SQLite as the shared datastore,
* explicit operational coordination for turn processing.

These choices may evolve, but any change must preserve the inward dependency rule and the single-source-of-truth rule for game behavior.

---

## Short Version

This project is one product with multiple executables.

* The **domain** defines the game.
* The **app** layer defines what the system does.
* **Infrastructure** handles SQLite, files, and auth details.
* **HTTP** and **CLI** are delivery mechanisms, not separate brains.
* **Runtime** wires everything together.
* The **frontend** presents server behavior; it does not own core rules.
* The **docs** explain the real system and must stay aligned with it.
