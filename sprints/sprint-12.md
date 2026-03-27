# Sprint 12: MVP Order Language

**Pass:** Pass 4
**Goal:** Define and implement parse-time order handling for the MVP order set: design the text grammar, produce typed `domain.Order` values, and expose an authenticated parse API without executing turns.
**Predecessor:** Sprint 11

---

## Sprint Rules

1. **One subsystem per task.** Each task targets exactly one bounded piece of
   work. If a task touches more than one subsystem, split it.

2. **Every task names its tests.** A task is not ready for an agent until it
   lists the exact tests to add or update.

3. **No unrelated cleanup.** Do not bundle opportunistic refactoring into a
   feature task. Required follow-through cleanup belongs in the same task
   when it is directly caused by the change: stale references, dead helpers,
   guard clauses, invariant fixes, API alignment within the touched
   subsystem, and tests.

4. **Tasks must fit in context.** Each task description must be self-contained:
   an agent should be able to read the task and begin work without needing to
   read the entire repository. Include file paths, function names, expected
   behavior, and acceptance criteria inline.

5. **Leave the repo green.** Every completed task must leave all existing tests
   passing. If a task would break an earlier pass, it is scoped wrong.

6. **Small diffs only.** Prefer several small tasks over one large one. If a
   task will touch more than ~200 lines or more than 3 files, split it.

7. **Every task must state failure paths and invariants.** If a task does
   lookup, selection, indexing, parsing, ID allocation, or file/template
   input, the task must define behavior for not-found / invalid / empty /
   duplicate cases and name tests for them.

8. **Every task must include an impact scan.** List the existing helpers,
   fields, comments, call sites, and tests that may become stale because of
   the change. Remove/update them in the same task, or explicitly say why
   they remain.

9. **New/changed APIs must match an existing pattern.** For ports, stores,
   constructors, and method signatures, the task must cite the existing
   pattern it follows. If it deviates, the task must briefly justify the
   deviation.

10. **Validation ownership must be explicit.** For external inputs
    (JSON/templates/CLI/API payloads), the task must say which SOUSA layer
    validates invariants (`domain` vs `app`) and what is validated.

---

## Context for Agents

Sprint 12 is **backend plus one internal design doc only**. Do not add frontend
work in this sprint. The Orders page parse button is a later sprint; this sprint
builds the backend contract that button will eventually call.

Today the system only stores and retrieves raw order text:

- `backend/internal/app/ports.go` defines `OrderStore`
- `backend/internal/infra/filestore/store.go` reads and writes `orders.txt`
- `backend/internal/delivery/http/handlers.go` exposes `GET/POST /api/:empireNo/orders`

There is **no order parser yet**, no typed `domain.Order` hierarchy, and no turn
execution. Sprint 12 stops at parse-time validation. It must not execute orders,
mutate `game.json`, or check runtime state such as whether a specific ship,
colony, or deposit currently exists.

### Parse-time scope for this sprint

- Command keyword recognition
- Per-command arity and clause shape
- Integer parsing and range checks that do not require game state
- Enum and unit-name validation against static domain types
- Distinguishing supported MVP orders from known-but-not-yet-implemented orders
- Returning all line diagnostics from one parse pass instead of failing fast

### Out of scope for this sprint

- Order execution
- Turn pipeline changes
- Colony/ship/deposit existence checks
- Ownership, reachability, and production-capacity validation
- Persisting parsed orders anywhere other than the existing in-memory return path

### Expected parsing behavior

- Input is raw text, processed line by line
- Blank lines are ignored
- Comment syntax is decided in Task 1 and then implemented consistently everywhere
- Valid lines still produce typed orders even if other lines fail
- Non-MVP historical commands return a `not_implemented` diagnostic, not a generic parse failure
- Move orders may preserve symbolic destinations even though the ship-orbit execution model is deferred to Sprint 16

### New API contract for this sprint

Add a protected endpoint:

```
POST /api/:empireNo/orders/parse
Authorization: Bearer <token>
Content-Type: text/plain
```

Response shape:

```json
{
  "ok": false,
  "accepted_count": 2,
  "diagnostics": [
    {
      "line": 3,
      "code": "not_implemented",
      "message": "bombard is not yet implemented"
    }
  ]
}
```

- `ok` is `true` only when `diagnostics` is empty
- `accepted_count` is the number of lines that produced typed `domain.Order` values
- The endpoint returns `200` for a successful parse pass even when diagnostics exist
- The endpoint returns `413` when the request body exceeds `maxOrderBytes`
- The endpoint returns `500` only for unexpected internal failures

### SOUSA split for Sprint 12

- `domain` owns typed order values and validation that does not need live game state
- `app` owns parse orchestration, diagnostics, and service contracts
- `infra` owns the concrete text parser implementation
- `delivery/http` owns request/response mapping only
- `runtime` is the only layer that instantiates the concrete parser and wires routes

After completing a task, update `sprints/sprint-12.md`: check off acceptance
criteria (change `[ ]` to `[x]`) and change the task status from TODO to DONE
in the Task Summary table at the bottom of the file.

**Key files:**
- `backend/internal/app/ports.go` — existing order-storage port pattern
- `backend/internal/domain/cluster.go` — existing typed IDs, enums, and unit kinds used by parsed orders
- `backend/internal/delivery/http/handlers.go` — existing orders handler patterns
- `backend/internal/delivery/http/routes.go` — protected route registration
- `backend/internal/runtime/server/server.go` — runtime wiring entry point
- `backend/internal/infra/filestore/store.go` — existing storage remains unchanged this sprint

**Likely new files:**
- `docs/order-language-v0.md`
- `backend/internal/domain/orders.go`
- `backend/internal/domain/orders_test.go`
- `backend/internal/app/order_parse_ports.go`
- `backend/internal/app/order_parse_service.go`
- `backend/internal/app/order_parse_service_test.go`
- `backend/internal/infra/ordertext/parser.go`
- `backend/internal/infra/ordertext/parser_test.go`

**Build/test commands:**
```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/api/
```

**Constraints reminder:**
- stdlib first — no parser generator and no new third-party dependency
- preserve the existing raw `orders.txt` storage API
- keep parsing deterministic and line-oriented
- prefer small tokenizer helpers over a speculative grammar framework

**Audit-left guidance:**
Audit items belong with the task that introduced the risk. SOUSA checks,
stale-reference scans, guard/negative tests, and API-pattern checks must
appear in the relevant task's design checklist or acceptance criteria. A
final audit task is only a last verification step — it should confirm
health, not discover new issues.

---

## Tasks

### Task 1: Order language design doc

**Subsystem:** internal design documentation
**Files:**
- `docs/order-language-v0.md` (new file)
**Depends on:** None

**What to do:**

Write the internal design doc that makes Sprint 12 implementation work concrete.
The document must define the v0 order text format, the canonical command names
and any aliases, the exact MVP command matrix, the proposed `domain.Order`
hierarchy, and the boundary between parse-time and execution-time validation.

The doc must explicitly cover these topics:

1. One-order-per-line grammar, including whitespace, blank lines, and comment handling.
2. Grammar and examples for every MVP order in the roadmap: Build Change, Mining Change,
   Draft, Pay, Ration, Transfer, Assemble variants, Set up, Move (in-system and system jump),
   and Name.
3. The `domain.Order` type family to create in Task 2, including how move destinations are
   represented without committing Sprint 16's ship-location execution model.
4. Which validations happen during parsing versus later execution, with at least one example
   of each category per major order family.
5. The phase number each order maps to, matching `sprints/roadmap-v0.md`.
6. How known non-MVP commands are reported as `not_implemented`.

This document is the authority for Tasks 2–6. Later tasks may not invent grammar
or diagnostics that conflict with it.

**Design review checklist:**

_SOUSA layers touched:_
- [x] domain
- [x] app
- [x] infra
- [x] delivery
- [x] runtime
- Allowed dependency direction: N/A — documentation only

_Existing pattern to follow:_
- `docs/SOUSA.md` — architectural vocabulary and boundary language
- `sprints/roadmap-v0.md` — MVP command set and phase mapping source of truth

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: N/A
- [x] Empty/nil/invalid input behavior specified: doc defines blank-line, comment, malformed-line, and unsupported-command behavior
- [x] ID/index bounds behavior specified (if applicable): numeric-field expectations called out per command

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: parse-time does not depend on live game state
- [x] Validation rules listed and layer assigned (`domain` or `app`)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `sprints/roadmap-v0.md` terminology only if the doc reveals a naming mismatch; otherwise none
- Search commands: `rg -n "MVP order set|order text syntax|parse-time|execution-time" sprints docs backend/internal`

**Acceptance criteria:**
- [x] `docs/order-language-v0.md` exists and covers grammar, type hierarchy, phase mapping, and validation ownership
- [x] Every MVP order has at least one concrete syntax example in the doc
- [x] Known non-MVP command handling is specified as `not_implemented`
- [x] Move-order syntax is defined in a way that can be parsed now without requiring Sprint 16 execution wiring
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `None` — documentation task
- Existing tests updated: None

---

### Task 2: Domain order model

**Subsystem:** `domain`
**Files:**
- `backend/internal/domain/orders.go` (new file)
- `backend/internal/domain/orders_test.go` (new file)
**Depends on:** Task 1

**What to do:**

Add the typed `domain.Order` model described in Task 1. Keep the shape minimal:
the domain owns typed order values and pure validation that does not require the
current game state.

Define concrete order types for the Sprint 12 MVP command set. Use existing typed
IDs and enums from `backend/internal/domain/cluster.go` where possible instead of
stringly-typed fields. If a command needs a destination or target shape that is not
yet represented elsewhere, add a small domain type for that purpose rather than
borrowing the future execution model.

The order types must support at least these capabilities:

- identify the order kind and phase
- preserve enough typed data for a later turn engine to execute the order
- validate pure invariants such as non-empty names, positive IDs/counts, valid
  percentage ranges, and valid enum/unit values

Do **not** add execution methods, store access, or references to `app`, `infra`,
`delivery`, or `runtime`.

**Design review checklist:**

_SOUSA layers touched:_
- [x] domain
- [x] app
- [x] infra
- [x] delivery
- [x] runtime
- Allowed dependency direction: domain only

_Existing pattern to follow:_
- `backend/internal/domain/cluster.go` — typed IDs, enums, and small pure structs
- `backend/internal/domain/game.go` — flat data structs with no framework imports

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: N/A — no live lookups in domain
- [x] Empty/nil/invalid input behavior specified: validation methods reject empty names, invalid percentages, invalid counts, and unknown enum-like values
- [x] ID/index bounds behavior specified (if applicable): IDs/counts must be positive; orbit/phase-adjacent values must be range-checked if represented numerically

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: parsed values are execution-ready but not execution-validated against game state
- [x] Validation rules listed and layer assigned (`domain` or `app`): domain validates only static invariants

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: any new order-kind enums or string methods referenced by parser or API response code
- Search commands: `rg -n "type .*ID|type .*Kind|type TechLevel|UnitKind" backend/internal/domain`

**Acceptance criteria:**
- [x] `backend/internal/domain/orders.go` exports typed order values for every MVP order family needed in Sprint 12
- [x] The domain model for move orders preserves parsed intent without requiring the later ship-location execution model
- [x] Pure validation methods exist for static invariants and do not require live game state
- [x] At least one negative/guard-path test added or updated (if task has lookups, parsing, or ID allocation)
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestOrdersValidate_HappyPath` in `backend/internal/domain/orders_test.go` — representative valid MVP orders pass pure validation
- `TestOrdersValidate_InvalidValues` in `backend/internal/domain/orders_test.go` — empty names, bad percentages, zero/negative IDs, and invalid enum-like values fail validation
- Existing tests updated: None

---

### Task 3: App parse contracts and service

**Subsystem:** `app`
**Files:**
- `backend/internal/app/order_parse_ports.go` (new file)
- `backend/internal/app/order_parse_service.go` (new file)
- `backend/internal/app/order_parse_service_test.go` (new file)
**Depends on:** Task 2

**What to do:**

Add the app-layer contract for parsing raw order text into typed domain orders.
Create an `OrderParser` port and a `ParseOrdersService` that delivery code can call.

The service should accept raw order text and return a stable app-layer result with:

- the successfully parsed `[]domain.Order`
- a per-line diagnostics slice containing at least `line`, `code`, and `message`

Expected behavior:

- empty body is valid and returns zero orders and zero diagnostics
- blank/comment lines do not create diagnostics
- valid and invalid lines may coexist in one response
- unexpected parser failures return an error instead of being turned into a fake line diagnostic

The service should not import `infra` or `delivery/http`. Keep response types in `app`
so delivery can JSON-encode them without knowing parser internals.

**Design review checklist:**

_SOUSA layers touched:_
- [x] domain
- [x] app
- [x] infra
- [x] delivery
- [x] runtime
- Allowed dependency direction: app → domain only

_Existing pattern to follow:_
- `backend/internal/app/services.go` — small service struct with one public method
- `backend/internal/app/template_ports.go` and `backend/internal/app/ports.go` — port interface placement and naming

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: N/A
- [x] Empty/nil/invalid input behavior specified: empty body succeeds; nil-equivalent parser failures bubble as errors
- [x] ID/index bounds behavior specified (if applicable): line numbers in diagnostics must be 1-based and stable

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: diagnostics remain in input order; accepted orders preserve input order
- [x] Validation rules listed and layer assigned (`domain` or `app`): app owns result packaging and diagnostic shape; domain owns pure order-value validation

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `backend/internal/delivery/http/handlers.go`, `routes.go`, and runtime wiring that will consume the new service
- Search commands: `rg -n "OrderStore|LoginService|DashboardStore|AddRoutes\(" backend/internal`

**Acceptance criteria:**
- [x] `OrderParser` port and `ParseOrdersService` exist in `app`
- [x] The parse result type contains typed orders plus stable diagnostics with `line`, `code`, and `message`
- [x] Empty input returns a successful empty result rather than an error
- [x] At least one negative/guard-path test added or updated (if task has lookups, parsing, or ID allocation)
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestParseOrdersService_EmptyInput` in `backend/internal/app/order_parse_service_test.go` — empty body yields zero orders and zero diagnostics
- `TestParseOrdersService_ReturnsOrdersAndDiagnostics` in `backend/internal/app/order_parse_service_test.go` — partial success preserves accepted orders and diagnostics in input order
- `TestParseOrdersService_PropagatesParserFailure` in `backend/internal/app/order_parse_service_test.go` — unexpected parser failure returns an error
- Existing tests updated: None

---

### Task 4: Infra text parser

**Subsystem:** `infra`
**Files:**
- `backend/internal/infra/ordertext/parser.go` (new file)
- `backend/internal/infra/ordertext/parser_test.go` (new file)
**Depends on:** Task 3

**What to do:**

Implement the concrete line-oriented text parser in a new infra package,
`backend/internal/infra/ordertext`. This package must satisfy the app-layer
`OrderParser` port from Task 3.

Behavior requirements:

- parse raw text line by line using the grammar locked in by Task 1
- build typed `domain.Order` values from valid MVP lines
- use domain validation for static invariant checks
- emit `not_implemented` for known historical commands outside the MVP set
- emit distinct diagnostics for malformed syntax versus invalid values
- continue parsing after bad lines so one request returns the full diagnostics set

Implementation constraints:

- no parser generator and no new dependency
- no file I/O and no reads from live game state
- helpers stay local to the package unless reused more than once for a real reason

**Design review checklist:**

_SOUSA layers touched:_
- [x] domain
- [x] app
- [x] infra
- [x] delivery
- [x] runtime
- Allowed dependency direction: infra → app, domain only

_Existing pattern to follow:_
- `backend/internal/infra/filestore/store.go` — concrete adapter with small constructor and no delivery imports
- `backend/internal/domain/cluster.go` string and enum patterns for unit/kind decoding

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: N/A — parser does not do live lookups
- [x] Empty/nil/invalid input behavior specified: blank/comment lines ignored; malformed lines reported with diagnostics; parser never panics on short tokens
- [x] ID/index bounds behavior specified (if applicable): numeric parsing errors and out-of-range static values become diagnostics with stable line numbers

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: diagnostics and accepted orders remain in input order
- [x] Validation rules listed and layer assigned (`domain` or `app`): infra tokenizes and maps text; domain validates static invariants; app packages results

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: new constructor wiring in runtime and new handler call sites in delivery
- Search commands: `rg -n "OrderParser|ParseOrdersService|not_implemented|orders/parse" backend/internal`

**Acceptance criteria:**
- [x] `backend/internal/infra/ordertext` satisfies the app `OrderParser` port
- [x] Representative examples for each MVP order family parse into typed domain orders
- [x] Blank lines and comment lines are ignored without diagnostics
- [x] Known non-MVP commands return `not_implemented` diagnostics instead of a generic unknown-command failure
- [x] At least one negative/guard-path test added or updated (if task has lookups, parsing, or ID allocation)
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestParser_ParseMVPOrders` in `backend/internal/infra/ordertext/parser_test.go` — representative valid MVP lines parse into the expected order kinds
- `TestParser_IgnoresBlankAndCommentLines` in `backend/internal/infra/ordertext/parser_test.go` — whitespace and comments produce no diagnostics
- `TestParser_RejectsUnsupportedAndMalformedLines` in `backend/internal/infra/ordertext/parser_test.go` — unsupported commands, missing fields, and invalid numeric values produce the right diagnostic codes
- Existing tests updated: None

---

### Task 5: HTTP order-parse endpoint

**Subsystem:** `delivery/http`
**Files:**
- `backend/internal/delivery/http/handlers.go`
- `backend/internal/delivery/http/handlers_test.go`
- `backend/internal/delivery/http/routes.go`
**Depends on:** Task 3

**What to do:**

Add the authenticated parse endpoint described in the sprint context:

```
POST /api/:empireNo/orders/parse
```

The handler should:

- require the existing JWT + empire ownership middleware path used by `GET/POST /api/:empireNo/orders`
- read `text/plain` request bodies using the same `maxOrderBytes` limit pattern as `PostOrders`
- call `ParseOrdersService`
- return JSON with `ok`, `accepted_count`, and `diagnostics`
- return `413` for oversized bodies and `500` for unexpected service errors

Keep the existing `GET/POST /api/:empireNo/orders` behavior unchanged.

**Design review checklist:**

_SOUSA layers touched:_
- [x] domain
- [x] app
- [x] infra
- [x] delivery
- [x] runtime
- Allowed dependency direction: delivery → app only

_Existing pattern to follow:_
- `backend/internal/delivery/http/handlers.go` — `PostOrders` request-body handling and error mapping
- `backend/internal/delivery/http/routes.go` — protected route registration under `EmpireAuthMiddleware`

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: N/A — parse endpoint does not look up stored orders
- [x] Empty/nil/invalid input behavior specified: empty body returns a successful empty parse; oversized body returns 413; service failure returns 500
- [x] ID/index bounds behavior specified (if applicable): line numbers are passed through from app diagnostics without rewriting

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: authenticated empire must match `:empireNo` via existing middleware; response preserves service diagnostic order
- [x] Validation rules listed and layer assigned (`domain` or `app`): delivery does HTTP/body validation only; parser semantics stay in app/infra/domain

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `AddRoutes` signature and any tests or comments that enumerate protected routes
- Search commands: `rg -n "PostOrders|GetOrders|AddRoutes\(|/orders" backend/internal/delivery/http`

**Acceptance criteria:**
- [x] `POST /api/:empireNo/orders/parse` is registered under the protected route group
- [x] The handler reuses the existing body-size limit pattern from `PostOrders`
- [x] A successful parse with diagnostics returns HTTP 200 and the documented JSON shape
- [x] At least one negative/guard-path test added or updated (if task has lookups, parsing, or ID allocation)
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestPostParseOrders_OK` in `backend/internal/delivery/http/handlers_test.go` — valid request returns 200 with `ok=true` and the accepted count
- `TestPostParseOrders_PartialSuccess` in `backend/internal/delivery/http/handlers_test.go` — diagnostics return 200 with `ok=false`
- `TestPostParseOrders_TooLarge` in `backend/internal/delivery/http/handlers_test.go` — oversized body returns 413
- `TestPostParseOrders_InternalError` in `backend/internal/delivery/http/handlers_test.go` — service failure returns 500
- Existing tests updated: any route-enumeration helpers in `backend/internal/delivery/http/handlers_test.go`, if needed

---

### Task 6: Runtime parser wiring

**Subsystem:** `runtime`
**Files:**
- `backend/internal/runtime/server/server.go`
- `backend/internal/runtime/server/server_test.go`
**Depends on:** Task 4, Task 5

**What to do:**

Wire the concrete parser into server startup. Runtime is the only layer that may
instantiate the infra parser and hand it to delivery via the app service.

Create the concrete parser with a small constructor in `infra/ordertext`, wrap it
in the `app.ParseOrdersService`, and pass that service into `delivery/http.AddRoutes`.
Do not let `delivery/http` import the infra package directly.

This task should be small and boring: it exists only to keep SOUSA wiring in the
correct layer.

**Design review checklist:**

_SOUSA layers touched:_
- [x] domain
- [x] app
- [x] infra
- [x] delivery
- [x] runtime
- Allowed dependency direction: runtime → app, infra, delivery

_Existing pattern to follow:_
- `backend/internal/runtime/server/server.go` — existing wiring of auth, filestore, and `LoginService`
- `backend/internal/delivery/http/routes.go` argument-passing pattern via `AddRoutes`

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: N/A
- [x] Empty/nil/invalid input behavior specified: parser construction is config-free and should not introduce new required startup config
- [x] ID/index bounds behavior specified (if applicable): N/A

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: runtime remains the only layer that imports the concrete parser package
- [x] Validation rules listed and layer assigned (`domain` or `app`): runtime performs no parser validation logic

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `delivery/http.AddRoutes` call sites and any startup smoke tests
- Search commands: `rg -n "AddRoutes\(|NewStore\(|LoginService|ordertext" backend/internal/runtime backend/internal/delivery`

**Acceptance criteria:**
- [x] `backend/internal/runtime/server/server.go` constructs the concrete text parser and wraps it in the app parse service
- [x] `delivery/http.AddRoutes` receives the new parse service from runtime rather than importing infra directly
- [x] Existing server startup behavior remains unchanged aside from the new route wiring
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestAutoShutdown` in `backend/internal/runtime/server/server_test.go` — existing startup smoke test still passes after parser wiring
- `TestNewMissingDeps` in `backend/internal/runtime/server/server_test.go` — existing constructor coverage remains valid with no new required config
- Existing tests updated: update only if route/service wiring changes startup helper assumptions

---

## Task Summary

| Task | Title                           | Status | Agent/Thread | Notes |
|------|---------------------------------|--------|--------------|-------|
| 1    | Order language design doc       | DONE   |              |       |
| 2    | Domain order model              | DONE   |              |       |
| 3    | App parse contracts and service | DONE   |              |       |
| 4    | Infra text parser               | DONE   |              |       |
| 5    | HTTP order-parse endpoint       | DONE   |              |       |
| 6    | Runtime parser wiring           | DONE   |              |       |
