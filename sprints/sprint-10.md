# Sprint 10: Dashboard API

**Pass:** Pass 3
**Goal:** Add a `GET /api/:empireNo/dashboard` endpoint that returns colony, ship, and planet summary counts for use by the Sprint 11 frontend dashboard.
**Predecessor:** Sprint 9

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

This sprint adds a single read-only API endpoint: `GET /api/:empireNo/dashboard`.
The response contains summary counts for the player's dashboard cards: colonies
by kind, ships (always 0 for now), and unique homeworld planets by kind.

The frontend is built in Sprint 11. This sprint is backend-only.

After completing a task, update `sprints/sprint-10.md`: check off acceptance
criteria (change `[ ]` to `[x]`) and change the task status from TODO to DONE
in the Task Summary table at the bottom of the file.

### API contract

```
GET /api/:empireNo/dashboard
Authorization: Bearer <token>

200 OK
{
  "colony_count":  1,
  "colony_kinds":  [{"kind": "Open Air", "count": 1}],
  "ship_count":    0,
  "planet_count":  1,
  "planet_kinds":  [{"kind": "Terrestrial", "count": 1}]
}
```

- `colony_kinds` and `planet_kinds` omit entries with count 0.
- `colony_count` is the sum of all `colony_kinds` counts.
- `planet_count` is the count of **unique** planets (by `PlanetID`) the
  empire has colonies on, grouped by `PlanetKind`.
- `ship_count` is always 0 until ships are implemented.
- If the empire is not found in `game.json`, the handler returns 404.
- Requires a valid JWT that matches the `:empireNo` path parameter
  (enforced by existing `EmpireAuthMiddleware`).

### Kind strings

Colony kind strings come from `domain.ColonyKind.String()`:
- `OpenAir` → `"Open Air"`
- `Orbital` → `"Orbital"`
- `Enclosed` → `"Enclosed"`

Planet kind strings come from `domain.PlanetKind.String()`:
- `Terrestrial` → `"Terrestrial"`
- `AsteroidBelt` → `"Asteroid Belt"`
- `GasGiant` → `"Gas Giant"`

### How `fileStore` knows the data path

`filestore.Store` is initialized with `dataPath` in `runtime/server/server.go`
at startup (`filestore.NewStore(s.dataPath)`). It holds `dataPath` as a field
and constructs file paths from it internally. The same `*Store` value is
already passed to `AddRoutes` as `orderStore` and `reportStore`. This sprint
adds it as a third store argument, `dashboardStore`.

### SOUSA constraints

- `app/dashboard_ports.go` — new port interface; imports only `domain`-free
  Go types (no `domain` import needed since `DashboardSummary` uses only
  `string` and `int`).
- `infra/filestore/dashboard.go` — implements the port; may import `domain`
  and Go stdlib only.
- `delivery/http/handlers.go` — new handler; imports `app` and `cerr` only,
  not `infra` or `domain` directly. Verify that `EmpireFromCtx` does not
  require a `domain` import in this file.
- `runtime/server/server.go` — the only file that imports `infra` and wires
  concrete types.

### Key files

- `backend/internal/app/dashboard_ports.go` — new file: port interface (Task 1)
- `backend/internal/infra/filestore/dashboard.go` — new file: implementation (Task 2)
- `backend/internal/infra/filestore/dashboard_test.go` — new file: tests (Task 2)
- `backend/internal/delivery/http/handlers.go` — new `GetDashboard` handler (Task 3)
- `backend/internal/delivery/http/routes.go` — new route + updated `AddRoutes` signature (Task 3)
- `backend/internal/runtime/server/server.go` — pass `fileStore` as `dashboardStore` (Task 3)

### Build/test commands

```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/api/
cd backend && go vet ./...
```

**Audit-left guidance:**
Audit items belong with the task that introduced the risk. SOUSA checks,
stale-reference scans, guard/negative tests, and API-pattern checks must
appear in the relevant task's design checklist or acceptance criteria. A
final audit task is only a last verification step — it should confirm
health, not discover new issues.

---

## Tasks

### Task 1: App — `DashboardSummary` type and `DashboardStore` port

**Subsystem:** `app`
**Files:**
- `backend/internal/app/dashboard_ports.go` (new file)
**Depends on:** None

**What to do:**

Create the port interface and response types for the dashboard endpoint.
These types are shared between the filestore implementation (Task 2) and the
HTTP handler (Task 3).

```go
// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

// KindCount pairs a human-readable kind label with a count.
type KindCount struct {
    Kind  string `json:"kind"`
    Count int    `json:"count"`
}

// DashboardSummary is the response payload for GET /api/:empireNo/dashboard.
type DashboardSummary struct {
    ColonyCount int         `json:"colony_count"`
    ColonyKinds []KindCount `json:"colony_kinds"`
    ShipCount   int         `json:"ship_count"`
    PlanetCount int         `json:"planet_count"`
    PlanetKinds []KindCount `json:"planet_kinds"`
}

// DashboardStore computes dashboard summary data for a given empire.
type DashboardStore interface {
    GetDashboardSummary(empireNo int) (DashboardSummary, error)
}
```

`KindCount` slices must omit entries where `Count == 0`.
`ColonyCount` is the sum of all `ColonyKinds` counts.
`PlanetCount` is the count of unique planet IDs across all colonies (not the
sum of `PlanetKinds` counts, though for v0 with one colony per planet these
are equivalent).

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [x] app (new port interface and types)
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: N/A — `app` types only, no imports

_Existing pattern to follow:_
- `app/template_ports.go` — port interface pattern with `Store` suffix
- `DashboardSummary` is a flat response struct like existing app types

_Failure paths / guard clauses:_
- [ ] Not-found behavior specified (N/A — pure type definitions)
- [ ] Empty/nil/invalid input behavior specified (N/A)
- [ ] ID/index bounds behavior specified (N/A)

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A
- [x] Ordering or state preconditions stated: `KindCount` slices omit zero counts, sorted by `Kind`
- [x] Validation rules listed and layer assigned: N/A — types only

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: None
- Search commands: N/A

**Acceptance criteria:**
- [x] `cd backend && go build ./...` succeeds
- [x] `DashboardStore` interface exists in `app/dashboard_ports.go`
- [x] `DashboardSummary` and `KindCount` types exist with correct JSON tags
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- None — interface and type definitions have no testable logic.

---

### Task 2: Filestore — implement `GetDashboardSummary`

**Subsystem:** `infra/filestore`
**Files:**
- `backend/internal/infra/filestore/dashboard.go` (new file)
- `backend/internal/infra/filestore/dashboard_test.go` (new file)
**Depends on:** Task 1

**What to do:**

Implement `GetDashboardSummary` on `*Store`. It reads `game.json` and
`cluster.json` from `s.dataPath`, finds the empire, and computes counts.

**1. `filestore/dashboard.go`:**

```go
func (s *Store) GetDashboardSummary(empireNo int) (app.DashboardSummary, error) {
    // implementation below
}

// compile-time interface check
var _ app.DashboardStore = (*Store)(nil)
```

**Algorithm:**

```
1. Read game.json:
   game, err := s.ReadGame(s.dataPath)
   Wrap error: "getDashboardSummary: %w"

2. Find the empire in game.Empires by int(empire.ID) == empireNo.
   If not found: return cerr.ErrNotFound wrapped as
   "getDashboardSummary: empire %d not found: %w"

3. Read cluster.json:
   cluster, err := s.ReadCluster(s.dataPath)
   Wrap error: "getDashboardSummary: %w"

4. Build a lookup map: colonyByID map[domain.ColonyID]domain.Colony
   by ranging over cluster.Colonies.

5. Build a lookup map: planetByID map[domain.PlanetID]domain.Planet
   by ranging over cluster.Planets.

6. Count colonies by kind:
   colonyCounts := map[string]int{}
   for _, colonyID := range empire.Colonies {
       col, ok := colonyByID[colonyID]
       if !ok { continue }  // data inconsistency — skip silently
       colonyCounts[col.Kind.String()]++
   }

7. Count unique planets by kind:
   seenPlanet := map[domain.PlanetID]bool{}
   planetCounts := map[string]int{}
   for _, colonyID := range empire.Colonies {
       col, ok := colonyByID[colonyID]
       if !ok { continue }
       if seenPlanet[col.Planet] { continue }
       seenPlanet[col.Planet] = true
       planet, ok := planetByID[col.Planet]
       if !ok { continue }
       planetCounts[planet.Kind.String()]++
   }

8. Convert maps to []KindCount slices (omit zero counts).
   Sort each slice by Kind ascending for deterministic output.

9. Assemble and return DashboardSummary:
   - ColonyCount: sum of colony kind counts
   - ColonyKinds: sorted []KindCount from colonyCounts
   - ShipCount: 0 (ships not yet implemented)
   - PlanetCount: len(seenPlanet)
   - PlanetKinds: sorted []KindCount from planetCounts
```

Use `"sort"` from stdlib to sort `[]KindCount` by `Kind` field ascending.

**2. `filestore/dashboard_test.go`:**

Tests use `t.TempDir()` and write minimal game.json / cluster.json files.
Do not use `filestore.Store` methods for writing — write raw JSON directly
with `os.WriteFile` to keep tests independent of the write path.

`TestGetDashboardSummary_OneColony`:
- Write game.json with one empire (ID=1) that has one colony (ID=1).
- Write cluster.json with one `Colony{ID:1, Planet:10, Kind: OpenAir}` and
  one `Planet{ID:10, Kind: Terrestrial}`.
- Call `GetDashboardSummary(1)`.
- Assert `ColonyCount == 1`, `ColonyKinds == [{Kind:"Open Air", Count:1}]`.
- Assert `ShipCount == 0`.
- Assert `PlanetCount == 1`, `PlanetKinds == [{Kind:"Terrestrial", Count:1}]`.

`TestGetDashboardSummary_MultipleKinds`:
- Empire has 2 colonies: one Open Air on a Terrestrial planet, one Orbital
  on a Gas Giant.
- Assert `ColonyKinds` has two entries; `PlanetKinds` has two entries.
- Assert entries are sorted by `Kind` ascending.

`TestGetDashboardSummary_DeduplicatesPlanets`:
- Empire has 2 colonies both on the same planet (ID=10).
- Assert `PlanetCount == 1`.

`TestGetDashboardSummary_EmpireNotFound`:
- game.json has no empire with ID=99.
- Assert error wraps `cerr.ErrNotFound`.

`TestGetDashboardSummary_NoColonies`:
- Empire exists but has no colonies.
- Assert `ColonyCount == 0`, `ColonyKinds == nil` or `[]`.
- Assert `PlanetCount == 0`, `ShipCount == 0`.

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain (imported for type access only)
- [ ] app (imports port type)
- [x] infra (`filestore/dashboard.go` implements `DashboardStore`)
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: infra → app, infra → domain

_Existing pattern to follow:_
- `filestore/templates.go` — `ReadGame`/`ReadCluster` call pattern on `*Store`
- Compile-time interface assertion: `var _ app.DashboardStore = (*Store)(nil)`

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: empire not in `game.json` → `cerr.ErrNotFound` (wrapped)
- [x] Empty/nil/invalid input behavior specified: no colonies → `ColonyCount == 0`, empty slices
- [x] ID/index bounds behavior specified: colony/planet ID lookup → skip silently on data inconsistency

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A — read-only
- [x] Ordering or state preconditions stated: kind slices sorted by `Kind` ascending
- [x] Validation rules listed and layer assigned: `infra` validates empire existence; `app` defines types

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: None — new file
- Search commands: `rg 'DashboardStore' backend/internal/`

**Acceptance criteria:**
- [x] `cd backend && go build ./...` succeeds
- [x] `cd backend && go test ./...` passes
- [x] `filestore.Store` satisfies `app.DashboardStore` (compile-time assertion)
- [x] Reads from `s.dataPath/game.json` and `s.dataPath/cluster.json`
- [x] Returns `cerr.ErrNotFound` (wrapped) when empire is not in game.json
- [x] `colony_kinds` and `planet_kinds` omit zero-count entries
- [x] Both kind slices are sorted by `Kind` ascending
- [x] Unique-planet deduplication is correct
- [x] `ShipCount` is always 0
- [x] All five tests pass
- [x] At least one negative/guard-path test added or updated (`TestGetDashboardSummary_EmpireNotFound`, `TestGetDashboardSummary_NoColonies`)
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestGetDashboardSummary_OneColony`
- `TestGetDashboardSummary_MultipleKinds`
- `TestGetDashboardSummary_DeduplicatesPlanets`
- `TestGetDashboardSummary_EmpireNotFound`
- `TestGetDashboardSummary_NoColonies`

---

### Task 3: Delivery, route, and runtime wiring

**Subsystem:** `delivery/http`, `runtime/server`
**Files:**
- `backend/internal/delivery/http/handlers.go`
- `backend/internal/delivery/http/routes.go`
- `backend/internal/runtime/server/server.go`
**Depends on:** Tasks 1, 2

**What to do:**

Add the `GetDashboard` handler, register the route, and wire `fileStore`
into the new parameter.

**1. Add `GetDashboard` to `delivery/http/handlers.go`:**

```go
// GetDashboard returns colony, ship, and planet summary counts for the
// authenticated empire.
// Requires EmpireAuthMiddleware to have validated ownership.
func GetDashboard(store app.DashboardStore) func(c *echo.Context) error {
    return func(c *echo.Context) error {
        empireNo, _ := EmpireFromCtx(c)

        summary, err := store.GetDashboardSummary(empireNo)
        if err != nil {
            if errors.Is(err, cerr.ErrNotFound) {
                return c.JSON(http.StatusNotFound, map[string]any{"error": "not found"})
            }
            return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
        }

        return c.JSON(http.StatusOK, summary)
    }
}
```

**2. Update `AddRoutes` in `delivery/http/routes.go`:**

Add `dashboardStore app.DashboardStore` as a new parameter after `reportStore`:

```go
func AddRoutes(
    e *echo.Echo,
    jwtMiddleware echo.MiddlewareFunc,
    empireExtractor EmpireExtractor,
    tokenValidator TokenValidator,
    loginSvc *app.LoginService,
    orderStore app.OrderStore,
    reportStore app.ReportStore,
    dashboardStore app.DashboardStore,  // new
    shutdownKey string,
    shutdownCh chan struct{},
    maxOrderBytes int64,
)
```

Register the new route in the protected group:

```go
protected.GET("/api/:empireNo/dashboard", GetDashboard(dashboardStore))
```

**3. Update `runtime/server/server.go`:**

Pass `fileStore` as the new `dashboardStore` argument in the `AddRoutes` call:

```go
deliveryhttp.AddRoutes(
    e,
    jwtMgr.Middleware(),
    empireExtractor,
    tokenValidator,
    loginSvc,
    fileStore,        // orderStore
    fileStore,        // reportStore
    fileStore,        // dashboardStore  ← new
    s.shutdownKey,
    s.shutdownCh,
    maxOrderBytes,
)
```

**4. Update `delivery/http/handlers_test.go`** if `AddRoutes` is called
there — update the call site to pass `nil` or a stub for `dashboardStore`.
Check whether `handlers_test.go` calls `AddRoutes` directly; if so, add a
`nil` argument in the correct position.

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [ ] app (imports port type)
- [ ] infra
- [x] delivery (`http/handlers.go`, `http/routes.go`)
- [x] runtime (`server/server.go` wiring)
- Allowed dependency direction: delivery → app; runtime → infra + delivery + app

_Existing pattern to follow:_
- `GetReport` handler pattern in `handlers.go` (closure returning handler func)
- `AddRoutes` parameter pattern in `routes.go`
- `server.go` wiring: same `fileStore` passed as multiple store interfaces

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: `cerr.ErrNotFound` → HTTP 404
- [x] Empty/nil/invalid input behavior specified: other errors → HTTP 500
- [ ] ID/index bounds behavior specified (N/A — `EmpireFromCtx` extracts empire number)

_Invariants / validation:_
- [x] Uniqueness / ID generation rule stated: N/A — read-only endpoint
- [x] Ordering or state preconditions stated: requires `EmpireAuthMiddleware` for auth
- [x] Validation rules listed and layer assigned: `delivery` — auth via middleware; `infra` — empire lookup

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `AddRoutes` call sites in `handlers_test.go` and `server.go`
- Search commands: `rg 'AddRoutes' backend/internal/`

**Acceptance criteria:**
- [x] `cd backend && go build ./...` succeeds
- [x] `cd backend && go test ./...` passes
- [x] `GET /api/:empireNo/dashboard` route is registered
- [x] `AddRoutes` accepts `dashboardStore app.DashboardStore`
- [x] `runtime/server/server.go` passes `fileStore` as `dashboardStore`
- [x] `GetDashboard` returns 200 with `DashboardSummary` JSON on success
- [x] `GetDashboard` returns 404 when empire not found
- [x] `GetDashboard` returns 500 on other errors
- [x] Any existing `AddRoutes` call sites (tests) compile after signature change
- [x] Stale references/helpers caused by this change removed or explicitly retained with reason
- [x] New/changed API matches an existing pattern (or deviation documented)
- [x] SOUSA boundary valid for touched layers

**Tests to add/update:**
- Update `delivery/http/handlers_test.go` call sites if `AddRoutes` is
  called there — add a `nil` dashboardStore argument.

---

### Task 4: Audit, build, and full test suite

**Subsystem:** all
**Files:** all files touched in Tasks 1–3
**Depends on:** Tasks 1–3

**What to do:**

1. **SOUSA import audit:**
   ```bash
   # delivery must not import infra or domain directly
   grep -r '"github.com/mdhender/ec/internal/infra"' backend/internal/delivery/
   grep -r '"github.com/mdhender/ec/internal/domain"' backend/internal/delivery/http/

   # app must not import infra or delivery
   grep -r '"github.com/mdhender/ec/internal/infra\|delivery"' backend/internal/app/
   ```
   Fix any violations.

2. **API contract check** — verify the JSON field names in `DashboardSummary`
   match the contract in this document (`colony_count`, `colony_kinds`, etc.).

3. **Zero-count omission** — verify that a `KindCount` with `Count == 0` is
   never present in the response. Check `TestGetDashboardSummary_NoColonies`.

4. **Run `go vet`:**
   ```bash
   cd backend && go vet ./...
   ```

5. **Full build and test:**
   ```bash
   cd backend && go build ./...
   cd backend && go test ./...
   ```

**Acceptance criteria:**
- [x] No SOUSA import violations
- [x] JSON field names match the contract
- [x] Zero-count entries never appear in kind slices
- [x] `cd backend && go vet ./...` passes
- [x] `cd backend && go build ./...` succeeds
- [x] `cd backend && go test ./...` passes

**Tests to add/update:**
- None — audit task only.

---

## Task Summary

| Task | Title                                         | Status | Depends On | Agent/Thread | Notes |
|------|-----------------------------------------------|--------|------------|--------------|-------|
| 1    | App: DashboardSummary + DashboardStore port   | DONE   | —          |              |       |
| 2    | Filestore: GetDashboardSummary + tests        | DONE   | 1          |              |       |
| 3    | Delivery + runtime: handler, route, wiring    | DONE   | 1, 2       |              |       |
| 4    | Audit, build, and full test suite             | DONE   | 1–3        |              |       |
