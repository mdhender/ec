# Sprint 8: Homeworld Template and Colony Foundation

**Pass:** Pass 2
**Goal:** Make homeworld setup fully template-driven and lay the domain foundation (types, ports, filestore) for colony seeding in Sprint 9.
**Predecessor:** Sprint 7

---

## Sprint Rules

1. **One subsystem per task.** Each task targets exactly one bounded piece of
   work. If a task touches more than one subsystem, split it.

2. **Every task names its tests.** A task is not ready for an agent until it
   lists the exact tests to add or update.

3. **No mixed concerns.** Never combine semantic translation with cleanup or
   refactoring in the same task.

4. **Tasks must fit in context.** Each task description must be self-contained:
   an agent should be able to read the task and begin work without needing to
   read the entire repository. Include file paths, function names, expected
   behavior, and acceptance criteria inline.

5. **Leave the repo green.** Every completed task must leave all existing tests
   passing. If a task would break an earlier pass, it is scoped wrong.

6. **Small diffs only.** Prefer several small tasks over one large one. If a
   task will touch more than ~200 lines or more than 3 files, split it.

---

## Context for Agents

This sprint has two concerns: (1) extending the domain to support colony
seeding (group types, updated Colony struct, template types), and (2) making
`CreateHomeWorld` template-driven so all homeworlds start with the same
deposits and habitability.

Colony seeding (populating inventory and groups in `AddEmpire`) is deferred to
Sprint 9. This sprint lays the foundation: types, ports, filestore, and the
homeworld template application.

After completing a task, update `sprints/sprint-8.md`: check off acceptance
criteria (change `[ ]` to `[x]`) and change the task status from TODO to DONE
in the Task Summary table at the bottom of the file.

### File layout

```
data-path/
  cluster.json             ← domain.Cluster
  game.json                ← domain.Game
  auth.json                ← domain.AuthConfig
  homeworld-template.json  ← domain.HomeworldTemplate  (NEW this sprint)
  colony-template.json     ← domain.ColonyTemplate     (NEW this sprint)
  1/                       ← empire subdirectory
  2/
  ...
```

Template files are placed by the GM before running setup commands. They are
read-only from the application's perspective.

### Architecture constraints

This project enforces the SOUSA layering rule: `domain ← app ← infra/delivery ← runtime`.

- `domain` — pure types and game rules, no framework imports
- `app` — use-case services and port interfaces; no SQL, HTTP, or file I/O
- `infra/filestore` — implements app ports using the filesystem
- `delivery/cli` — thin command wrappers; calls app services, no game logic
- `runtime/cli` — wires concrete types; the only layer that knows about infra

`delivery` must not import `infra`. `runtime` is the only layer that
instantiates and injects concrete implementations.

### Ordering invariant

`CreateHomeWorld` must be called for a planet before `AddEmpire` can assign
empires to it. This is enforced by the race lookup in `AddEmpire`: if no
`domain.Race` exists with `HomeWorld == homeWorldID`, the call fails. This
invariant must remain intact after all changes in this sprint.

### Key files

- `backend/internal/domain/cluster.go` — `Colony`, `Inventory`, `Deposit`, unit/group types (Tasks 1, 2)
- `backend/internal/domain/templates.go` — new file: `HomeworldTemplate`, `ColonyTemplate` (Task 3)
- `backend/internal/app/game_service.go` — `GameService`, `CreateHomeWorld`, `AddEmpire` (Tasks 2, 6)
- `backend/internal/app/game_service_test.go` — all service tests (Tasks 2, 6)
- `backend/internal/app/template_ports.go` — new file: `TemplateStore` interface (Task 4)
- `backend/internal/infra/filestore/templates.go` — new file: template readers (Task 5)
- `backend/internal/infra/filestore/templates_test.go` — new file: template round-trip tests (Task 5)
- `backend/internal/runtime/cli/cli.go` — wiring (Task 6)

### Key types (before this sprint)

```go
// domain/cluster.go
type Colony struct {
    ID        ColonyID
    Empire    EmpireID
    Location  Coords     // REMOVED in Task 2
    TechLevel TechLevel
}

type Inventory struct {
    Unit              UnitKind
    TechLevel         TechLevel
    QuantityAssembled int
    // QuantityDisassembled added in Task 1
}

// app/game_service.go
type GameService struct {
    Store   GameStore
    Cluster ClusterStore
    // Templates TemplateStore added in Task 6
}
```

### Build/test commands

```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/api/
cd backend && go build ./cmd/cli/
cd backend && go vet ./...
```

---

## Tasks

### Task 1: Domain — additive type additions

**Subsystem:** `domain`
**Files:**
- `backend/internal/domain/cluster.go`
**Depends on:** None

**What to do:**

Add new types and extend `Inventory`. All changes are additive — nothing is
removed or renamed, so no existing code breaks.

1. Add `QuantityDisassembled int` to `Inventory`:

   ```go
   type Inventory struct {
       Unit                UnitKind
       TechLevel           TechLevel
       QuantityAssembled   int
       QuantityDisassembled int
   }
   ```

2. Add `ColonyKind` type and constants:

   ```go
   type ColonyKind int

   const (
       OpenAir  ColonyKind = iota + 1
       Orbital
       Enclosed
   )

   func (k ColonyKind) String() string {
       switch k {
       case OpenAir:
           return "Open Air"
       case Orbital:
           return "Orbital"
       case Enclosed:
           return "Enclosed"
       default:
           return "Unknown"
       }
   }
   ```

3. Add group types:

   ```go
   // GroupUnit is a sub-group of a colony or ship group, representing
   // all units of the same tech level assigned to that group.
   type GroupUnit struct {
       TechLevel TechLevel
       Quantity  int
   }

   type MiningGroupID int

   type MiningGroup struct {
       ID      MiningGroupID
       Deposit DepositID
       Units   []GroupUnit
   }

   type FarmGroupID int

   // FarmGroup represents all farming units on a colony.
   // Each colony has at most one FarmGroup; sub-groups are by tech level.
   type FarmGroup struct {
       ID    FarmGroupID
       Units []GroupUnit
   }

   type FactoryGroupID int

   type FactoryGroup struct {
       ID    FactoryGroupID
       Units []GroupUnit
   }
   ```

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `Inventory` has `QuantityDisassembled int`
- [ ] `ColonyKind` type exists with `OpenAir`, `Orbital`, `Enclosed` constants
- [ ] `MiningGroup`, `FarmGroup`, `FactoryGroup`, `GroupUnit` types exist
- [ ] All existing tests still pass (additive change only)

**Tests to add/update:**
- None — pure additive type definitions with no logic.

---

### Task 2: Domain — update `Colony`; fix `AddEmpire` compilation

**Subsystem:** `domain`, `app/game_service`
**Files:**
- `backend/internal/domain/cluster.go`
- `backend/internal/app/game_service.go`
**Depends on:** Task 1

**What to do:**

Replace `Colony.Location Coords` with `Colony.Planet PlanetID` and add the
new group and inventory fields. Then fix the one call site in `AddEmpire`
that breaks.

1. **`domain/cluster.go` — update `Colony`:**

   ```go
   type Colony struct {
       ID            ColonyID
       Empire        EmpireID
       Planet        PlanetID        // replaces Location
       Kind          ColonyKind
       TechLevel     TechLevel
       Inventory     []Inventory
       MiningGroups  []MiningGroup
       FarmGroups    []FarmGroup
       FactoryGroups []FactoryGroup
   }
   ```

   `Location Coords` is **removed**. `Planet PlanetID` is the new locator.
   Colonies never move, so system coordinates are looked up from the planet
   at report time.

2. **`app/game_service.go` — fix `AddEmpire`:**

   - Remove the `findSystemForPlanet` call (currently lines ~112–115):
     ```go
     // DELETE these lines:
     systemLocation, err := findSystemForPlanet(cluster, homeWorldID)
     if err != nil {
         return 0, "", "", fmt.Errorf("addEmpire: %w", err)
     }
     ```
   - Update the colony construction to use `Planet` instead of `Location`:
     ```go
     colony := domain.Colony{
         ID:     domain.ColonyID(len(cluster.Colonies) + 1),
         Empire: domain.EmpireID(empireNo),
         Planet: homeWorldID,   // was: Location: systemLocation
     }
     ```
   - Add an explanatory comment above the ordering invariant check (the race
     lookup). Place it just before the `raceIdx` loop:
     ```go
     // Ordering invariant: CreateHomeWorld must have been called for this
     // planet before AddEmpire can assign empires to it. The race lookup
     // below enforces this — if no Race exists with HomeWorld == homeWorldID,
     // the operation fails.
     ```

   Do **not** delete `findSystemForPlanet` — it is still used by
   `CreateHomeWorld`.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `Colony` has `Planet PlanetID`; `Location Coords` does not exist
- [ ] `Colony` has `Kind`, `TechLevel`, `Inventory`, `MiningGroups`, `FarmGroups`, `FactoryGroups` fields
- [ ] `AddEmpire` sets `colony.Planet = homeWorldID`
- [ ] `findSystemForPlanet` still exists (used by `CreateHomeWorld`)
- [ ] Ordering invariant comment is present above the race lookup

**Tests to add/update:**
- No test changes needed — existing tests do not construct `Colony` directly
  and `AddEmpire` tests do not assert on `colony.Location`. Verify
  `go test ./...` still passes.

---

### Task 3: Domain — template types

**Subsystem:** `domain`
**Files:**
- `backend/internal/domain/templates.go` (new file)
**Depends on:** Task 1

**What to do:**

Create the domain types used to represent homeworld and colony setup
templates. These are read from JSON files by the filestore layer (Task 5).

```go
// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

// DepositTemplate describes a single natural resource deposit to create
// on a homeworld planet.
type DepositTemplate struct {
    Resource          NaturalResource
    YieldPct          int
    QuantityRemaining int
}

// HomeworldTemplate defines the starting conditions applied to a planet
// when it is designated as a homeworld. All homeworlds start with the
// same deposits and habitability.
type HomeworldTemplate struct {
    Habitability int
    Deposits     []DepositTemplate
}

// ColonyTemplate defines the starting state of the colony created when
// an empire is assigned to a homeworld.
type ColonyTemplate struct {
    Kind      ColonyKind
    TechLevel TechLevel
    Inventory []Inventory
}
```

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `domain.HomeworldTemplate`, `domain.ColonyTemplate`, `domain.DepositTemplate` exist
- [ ] `HomeworldTemplate` has `Habitability int` and `Deposits []DepositTemplate`
- [ ] `ColonyTemplate` has `Kind ColonyKind`, `TechLevel TechLevel`, `Inventory []Inventory`

**Tests to add/update:**
- None — pure type definitions with no logic.

---

### Task 4: App — `TemplateStore` port interface

**Subsystem:** `app`
**Files:**
- `backend/internal/app/template_ports.go` (new file)
**Depends on:** Task 3

**What to do:**

Define the port interface for reading template files. This follows the same
pattern as `GameStore` and `ClusterStore` in `app/game_ports.go` and
`app/cluster_ports.go`.

```go
// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// TemplateStore reads setup templates from the data directory.
// Template files are read-only; this interface has no write methods.
type TemplateStore interface {
    // ReadHomeworldTemplate reads homeworld-template.json from dataPath.
    ReadHomeworldTemplate(dataPath string) (domain.HomeworldTemplate, error)
    // ReadColonyTemplate reads colony-template.json from dataPath.
    ReadColonyTemplate(dataPath string) (domain.ColonyTemplate, error)
}
```

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `TemplateStore` interface exists in `app/template_ports.go`
- [ ] Interface has `ReadHomeworldTemplate` and `ReadColonyTemplate` methods

**Tests to add/update:**
- None — interface definitions have no testable logic.

---

### Task 5: Filestore — implement template readers

**Subsystem:** `infra/filestore`
**Files:**
- `backend/internal/infra/filestore/templates.go` (new file)
- `backend/internal/infra/filestore/templates_test.go` (new file)
**Depends on:** Tasks 3, 4

**What to do:**

Implement `ReadHomeworldTemplate` and `ReadColonyTemplate` on `*Store`.
Both read fixed filenames from `dataPath` and unmarshal JSON.

1. **`filestore/templates.go`:**

   ```go
   func (s *Store) ReadHomeworldTemplate(dataPath string) (domain.HomeworldTemplate, error) {
       // read filepath.Join(dataPath, "homeworld-template.json")
       // unmarshal into domain.HomeworldTemplate
       // return clear error if file does not exist
   }

   func (s *Store) ReadColonyTemplate(dataPath string) (domain.ColonyTemplate, error) {
       // read filepath.Join(dataPath, "colony-template.json")
       // unmarshal into domain.ColonyTemplate
       // return clear error if file does not exist
   }
   ```

   Verify that `*filestore.Store` satisfies `app.TemplateStore` by adding a
   compile-time assertion in this file:
   ```go
   var _ app.TemplateStore = (*Store)(nil)
   ```

   Follow the same JSON read pattern used in `filestore/cluster.go` and
   `filestore/game.go` (`os.ReadFile` + `json.Unmarshal`).

2. **`filestore/templates_test.go`:** Two round-trip tests using `t.TempDir()`.

   - `TestReadHomeworldTemplate`: write a `homeworld-template.json` with known
     values (`Habitability: 30`, two deposits), call `ReadHomeworldTemplate`,
     assert all fields match.
   - `TestReadColonyTemplate`: write a `colony-template.json` with known
     values (`Kind: OpenAir`, `TechLevel: 1`, two inventory entries), call
     `ReadColonyTemplate`, assert all fields match.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `filestore.Store` satisfies `app.TemplateStore` (compile-time assertion)
- [ ] `ReadHomeworldTemplate` reads `<dataPath>/homeworld-template.json`
- [ ] `ReadColonyTemplate` reads `<dataPath>/colony-template.json`
- [ ] Both return a clear error when the file does not exist
- [ ] `TestReadHomeworldTemplate` and `TestReadColonyTemplate` pass

**Tests to add/update:**
- `TestReadHomeworldTemplate` in `backend/internal/infra/filestore/templates_test.go`
- `TestReadColonyTemplate` in `backend/internal/infra/filestore/templates_test.go`

---

### Task 6: Update `CreateHomeWorld` to apply homeworld template; wire `TemplateStore`

**Subsystem:** `app/game_service`, `runtime/cli`
**Files:**
- `backend/internal/app/game_service.go`
- `backend/internal/app/game_service_test.go`
- `backend/internal/runtime/cli/cli.go`
**Depends on:** Tasks 2, 4, 5

**What to do:**

Add `Templates TemplateStore` to `GameService` and update `CreateHomeWorld`
to read and apply the homeworld template instead of hardcoding habitability
and leaving existing deposits in place.

**1. Update `GameService` struct:**

```go
type GameService struct {
    Store     GameStore
    Cluster   ClusterStore
    Templates TemplateStore
}
```

**2. Update `CreateHomeWorld` in `game_service.go`:**

After the planet is selected or validated (and before writing files), apply
the homeworld template:

```
a. Load template:
   tmpl, err := s.Templates.ReadHomeworldTemplate(dataPath)
   // wrap error: "createHomeWorld: %w"

b. Find the planet's current max deposit ID across the whole cluster
   (walk cluster.Deposits, track max int(d.ID)).

c. Find the planet by index in cluster.Planets (loop by ID, not slice index).
   On the planet:
   - Collect old DepositIDs from planet.Deposits.
   - Remove those Deposit records from cluster.Deposits (filter the slice).
   - Clear planet.Deposits to nil.

d. For each dt in tmpl.Deposits:
   nextID := maxDepositID + 1  (increment after each use)
   append domain.Deposit{ID: domain.DepositID(nextID), Resource: dt.Resource,
       YieldPct: dt.YieldPct, QuantityRemaining: dt.QuantityRemaining}
   to cluster.Deposits.
   append domain.DepositID(nextID) to planet.Deposits.

e. Set planet.Habitability = tmpl.Habitability
   (replaces the hardcoded `cluster.Planets[i].Habitability = 25`).
```

All planet mutations must go through the slice index (e.g.,
`cluster.Planets[i].Habitability = ...`), not a local copy, because Go
slice elements are not addressable via range value.

**3. Update `runtime/cli/cli.go`:**

Add `Templates: store` to the `gameSvc` literal:

```go
gameSvc := &app.GameService{
    Store:     store,
    Cluster:   store,
    Templates: store,
}
```

**4. Add `mockTemplateStore` to `game_service_test.go`:**

```go
type mockTemplateStore struct {
    homeworldTemplate domain.HomeworldTemplate
    colonyTemplate    domain.ColonyTemplate
    forceErr          error
}

func (m *mockTemplateStore) ReadHomeworldTemplate(dataPath string) (domain.HomeworldTemplate, error) {
    if m.forceErr != nil {
        return domain.HomeworldTemplate{}, m.forceErr
    }
    return m.homeworldTemplate, nil
}

func (m *mockTemplateStore) ReadColonyTemplate(dataPath string) (domain.ColonyTemplate, error) {
    if m.forceErr != nil {
        return domain.ColonyTemplate{}, m.forceErr
    }
    return m.colonyTemplate, nil
}
```

Add a helper `defaultHomeworldTemplate()` that returns a
`domain.HomeworldTemplate` with `Habitability: 25` and one deposit
(`FUEL`, `YieldPct: 50`, `QuantityRemaining: 100`). Use this in all
existing `TestCreateHomeWorld_*` tests by wiring
`Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}`
into the `GameService`.

**5. Update all existing `TestCreateHomeWorld_*` tests:**

Add `Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}`
to each `GameService` literal. No other changes are needed to these tests
unless they assert on `Habitability` (update those assertions to match
`defaultHomeworldTemplate().Habitability`).

**6. Add `TestCreateHomeWorld_AppliesTemplate`:**

Verify the template is applied correctly:
- Build a cluster with planet ID 1 that has two pre-existing deposits (IDs 10, 11).
- Provide a homeworld template with `Habitability: 30` and two deposits
  (METALLICS 60%, NONMETALLICS 40%).
- Call `CreateHomeWorld(dir, 1, 0)`.
- Assert:
  - `planet.Habitability == 30`
  - `planet.Deposits` has exactly 2 entries with IDs > 11 (old IDs deleted, new ones generated)
  - `cluster.Deposits` contains the two new deposits with correct Resource/YieldPct
  - The old deposit records (IDs 10, 11) are gone from `cluster.Deposits`

**7. Add `TestCreateHomeWorld_TemplateError`:**

Set `mockTemplateStore.forceErr` to a non-nil error. Call `CreateHomeWorld`.
Assert it returns an error (template read failure propagates).

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `GameService` has a `Templates TemplateStore` field
- [ ] `CreateHomeWorld` reads the homeworld template via `s.Templates`
- [ ] Habitability is set from the template, not hardcoded
- [ ] Existing deposits on the planet are deleted before applying the template
- [ ] New deposits get fresh global `DepositID`s (> any existing deposit ID)
- [ ] `runtime/cli/cli.go` passes `Templates: store` to `GameService`
- [ ] `TestCreateHomeWorld_AppliesTemplate` passes
- [ ] `TestCreateHomeWorld_TemplateError` passes
- [ ] All previously passing `TestCreateHomeWorld_*` tests still pass

**Tests to add/update:**
- `mockTemplateStore` + `defaultHomeworldTemplate()` — new helpers in `game_service_test.go`
- Update all `TestCreateHomeWorld_*` — add `Templates` to service constructor
- `TestCreateHomeWorld_AppliesTemplate` — verifies deposit replacement and habitability
- `TestCreateHomeWorld_TemplateError` — verifies template read error propagates

---

### Task 7: Audit, build, and full test suite

**Subsystem:** all
**Files:** all files touched in Tasks 1–6
**Depends on:** Tasks 1–6

**What to do:**

1. **SOUSA import audit** — verify no layering violations were introduced:
   ```bash
   # domain must not import app, infra, delivery, or runtime
   grep -r '"github.com/mdhender/ec/internal/app\|infra\|delivery\|runtime"' backend/internal/domain/

   # app must not import infra, delivery, runtime, or framework packages
   grep -r '"github.com/mdhender/ec/internal/infra\|delivery\|runtime"' backend/internal/app/

   # delivery must not import infra
   grep -r '"github.com/mdhender/ec/internal/infra"' backend/internal/delivery/
   ```
   Fix any violations found.

2. **Check for stale `Location` references** — verify `Colony.Location` is
   gone everywhere:
   ```bash
   grep -r "\.Location" backend/internal/ | grep -i colony
   ```
   Fix any found.

3. **Ordering invariant** — verify the comment added in Task 2 is present in
   `game_service.go` above the race lookup in `AddEmpire`.

4. **Run `go vet`:**
   ```bash
   cd backend && go vet ./...
   ```
   Fix any warnings.

5. **Run full build and test suite:**
   ```bash
   cd backend && go build ./...
   cd backend && go test ./...
   ```

**Acceptance criteria:**
- [ ] No SOUSA import violations
- [ ] No `Colony.Location` references remain
- [ ] Ordering invariant comment is present in `AddEmpire`
- [ ] `cd backend && go vet ./...` passes
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes

**Tests to add/update:**
- None — audit task only.

---

## Task Summary

| Task | Title                                                  | Status | Depends On | Agent/Thread | Notes |
|------|--------------------------------------------------------|--------|------------|--------------|-------|
| 1    | Domain: additive type additions                        | TODO   | —          |              |       |
| 2    | Domain: update Colony + fix AddEmpire                  | TODO   | 1          |              |       |
| 3    | Domain: template types                                 | TODO   | 1          |              |       |
| 4    | App: TemplateStore port interface                      | TODO   | 3          |              |       |
| 5    | Filestore: implement template readers                  | TODO   | 3, 4       |              |       |
| 6    | Update CreateHomeWorld + runtime wiring                | TODO   | 2, 4, 5    |              |       |
| 7    | Audit, build, and full test suite                      | TODO   | 1–6        |              |       |
