# Sprint 9: Colony Seeding

**Pass:** Pass 2
**Goal:** Seed the starting colony when an empire is created — populate it from the colony template, create its farm group, and assign mining groups to deposits.
**Predecessor:** Sprint 8

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

This sprint completes the empire creation workflow by fully seeding the
starting colony. All three tasks modify `AddEmpire` in `app/game_service.go`
and must be done in sequence.

**No CLI changes are needed.** The `CmdAddEmpire` command and its flags are
unchanged — template reading is internal to `GameService`.

**No new port interfaces or filestore code are needed.** `TemplateStore` and
`ReadColonyTemplate` were added in Sprint 8. This sprint only adds logic to
`AddEmpire` that calls the already-wired `s.Templates.ReadColonyTemplate`.

After completing a task, update `sprints/sprint-9.md`: check off acceptance
criteria (change `[ ]` to `[x]`) and change the task status from TODO to DONE
in the Task Summary table at the bottom of the file.

### How groups relate to inventory

Inventory records what unit quantities exist on the colony.
Groups record what is assigned to a specific purpose (mine, farm, factory).
The sum of all `MiningGroup` unit quantities must equal the total assembled
`Mine` units in inventory. This invariant is established at colony creation
and enforced by future turn-processing code — not in this sprint.

`FactoryGroups` are **not** assigned during empire creation; that is the
player's job via setup orders. The `FactoryGroups` field on `Colony` exists
but is left nil.

### Group and sub-group IDs

Group IDs (`MiningGroupID`, `FarmGroupID`) are **per-colony**, not global
across the cluster. Number them sequentially starting at 1 within each
colony. `MiningGroup` IDs run 1..N where N is the number of deposits.
The single `FarmGroup` always has ID 1.

### Mining group algorithm — "simplest thing that could possibly work"

> **Note to agents:** This algorithm is intentionally minimal. Future sprints
> will rework mining group assignment. Do not over-engineer.

1. Collect all `Mine` inventory entries where `QuantityAssembled > 0`,
   sorted by `TechLevel` ascending. Call this the *mine pool*.
2. `N` = number of deposits on the homeworld planet
   (`len(planet.Deposits)` where the planet is found by `Planet.ID == homeWorldID`
   in `cluster.Planets`).
3. If `N == 0` or mine pool is empty, set `colony.MiningGroups = nil` and skip.
4. `total` = sum of `QuantityAssembled` across all mine pool entries.
5. `base = total / N`, `remainder = total % N`.
6. For each deposit at index `i` (using `planet.Deposits[i]` as `depositID`):
   - `groupQty = base + 1` if `i < remainder`, else `base`.
   - Consume `groupQty` units from the mine pool greedily (deplete TL1 first,
     then TL2, etc.), collecting `[]GroupUnit` sub-groups as you go.
   - Append `MiningGroup{ID: MiningGroupID(i+1), Deposit: depositID, Units: subGroups}`
     to `colony.MiningGroups`.

The mine pool is a working copy; do **not** modify `colony.Inventory`.

### Farm group algorithm

1. Collect all `Farm` inventory entries where `QuantityAssembled > 0`,
   as `[]GroupUnit` grouped by `TechLevel` ascending.
2. If no Farm units, set `colony.FarmGroups = nil` and skip.
3. Otherwise: `colony.FarmGroups = []domain.FarmGroup{{ID: 1, Units: units}}`.

### Test infrastructure from Sprint 8

`game_service_test.go` already has:
- `mockTemplateStore` with `homeworldTemplate` and `colonyTemplate` fields
- `mockClusterStore`
- `makeTestCluster(planetID)` — builds a cluster with one system, one star,
  one terrestrial planet; **no deposits**

Sprint 9 adds `defaultColonyTemplate()` and `makeTestClusterWithDeposits()`
helpers (defined in Task 1). All new and updated tests use these.

### Architecture constraints

SOUSA layering: `domain ← app ← infra/delivery ← runtime`.
`app/game_service.go` may only import `domain` and Go stdlib. No infra,
no delivery, no framework packages.

### Key files

- `backend/internal/app/game_service.go` — `AddEmpire` (all three tasks)
- `backend/internal/app/game_service_test.go` — tests (all three tasks)
- `backend/internal/domain/cluster.go` — `Colony`, `MiningGroup`, `FarmGroup`, `GroupUnit`, `Mine`, `Farm` constants

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

### Task 1: Update `AddEmpire` — create colony from colony template

**Subsystem:** `app/game_service`
**Files:**
- `backend/internal/app/game_service.go`
- `backend/internal/app/game_service_test.go`
**Depends on:** None (Sprint 8 must be complete)

**What to do:**

Replace the minimal colony construction in `AddEmpire` with one that reads
the colony template and copies `Kind`, `TechLevel`, and `Inventory` from it.
Group creation is deferred to Tasks 2 and 3.

**1. Update `AddEmpire` in `game_service.go`:**

Add a template read immediately after the cluster is loaded (after the
`s.Cluster.ReadCluster` call):

```go
colonyTmpl, err := s.Templates.ReadColonyTemplate(dirPath)
if err != nil {
    return 0, "", "", fmt.Errorf("addEmpire: %w", err)
}
```

Replace the existing colony construction block with:

```go
// Deep-copy inventory from template so each empire gets independent slices.
inventory := make([]domain.Inventory, len(colonyTmpl.Inventory))
copy(inventory, colonyTmpl.Inventory)

colony := domain.Colony{
    ID:        domain.ColonyID(len(cluster.Colonies) + 1),
    Empire:    domain.EmpireID(empireNo),
    Planet:    homeWorldID,
    Kind:      colonyTmpl.Kind,
    TechLevel: colonyTmpl.TechLevel,
    Inventory: inventory,
}
```

`colony.MiningGroups`, `colony.FarmGroups`, and `colony.FactoryGroups` are
left nil; they are populated in Tasks 2 and 3.

**2. Add test helpers to `game_service_test.go`:**

Add `defaultColonyTemplate()`:

```go
func defaultColonyTemplate() domain.ColonyTemplate {
    return domain.ColonyTemplate{
        Kind:      domain.OpenAir,
        TechLevel: 1,
        Inventory: []domain.Inventory{
            {Unit: domain.Farm,    TechLevel: 1, QuantityAssembled: 10},
            {Unit: domain.Mine,    TechLevel: 1, QuantityAssembled: 20},
            {Unit: domain.Factory, TechLevel: 1, QuantityAssembled: 5},
        },
    }
}
```

Add `makeTestClusterWithDeposits(planetID, depositIDs)`:

```go
// makeTestClusterWithDeposits builds a cluster with one system, one star,
// one terrestrial planet that has the given deposits pre-populated.
func makeTestClusterWithDeposits(planetID domain.PlanetID, depositIDs []domain.DepositID) domain.Cluster {
    deposits := make([]domain.Deposit, len(depositIDs))
    for i, id := range depositIDs {
        deposits[i] = domain.Deposit{
            ID:                id,
            Resource:          domain.METALLICS,
            YieldPct:          50,
            QuantityRemaining: 1000,
        }
    }
    return domain.Cluster{
        Systems: []domain.System{
            {ID: 1, Location: domain.Coords{X: 10, Y: 10, Z: 10}},
        },
        Stars: []domain.Star{
            {ID: 1, System: 1, Orbits: [10]domain.PlanetID{planetID}},
        },
        Planets: []domain.Planet{
            {ID: planetID, Kind: domain.Terrestrial, Habitability: 25, Deposits: depositIDs},
        },
        Deposits: deposits,
    }
}
```

**3. Update all existing `TestAddEmpire_*` tests** that construct `GameService`
to include `Templates`:

```go
svc := &app.GameService{
    Store:     store,
    Cluster:   clusterStore,
    Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()},
}
```

Apply this to: `TestAddEmpire` (all sub-tests), `TestAddEmpire_RequiresHomeWorld`,
`TestAddEmpire_HomeWorldOverride`, `TestAddEmpire_HomeWorldNotFound`,
`TestAddEmpire_HomeWorldFull`, `TestScrubEmpireName`.

**4. Add `TestAddEmpire_ColonyFromTemplate`:**

- Set up a game with a homeworld at planet 100 and a `defaultColonyTemplate()`.
- Call `AddEmpire`.
- Read back the cluster from `mockClusterStore`.
- Assert `len(cluster.Colonies) == 1`.
- Assert `colony.Planet == 100`.
- Assert `colony.Kind == domain.OpenAir`.
- Assert `colony.TechLevel == 1`.
- Assert `len(colony.Inventory) == 3` (matches template).
- Assert the Inventory slice is a copy, not the same slice as the template
  (modify one, verify the other is unchanged).

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `AddEmpire` reads the colony template via `s.Templates.ReadColonyTemplate`
- [ ] Colony `Kind` and `TechLevel` are copied from the template
- [ ] Colony `Inventory` is a deep copy of the template inventory
- [ ] Colony `MiningGroups`, `FarmGroups`, `FactoryGroups` are nil (not yet assigned)
- [ ] All updated `TestAddEmpire_*` tests pass with `Templates` wired
- [ ] `TestAddEmpire_ColonyFromTemplate` passes

**Tests to add/update:**
- `defaultColonyTemplate()` helper — new in `game_service_test.go`
- `makeTestClusterWithDeposits()` helper — new in `game_service_test.go`
- Update all `TestAddEmpire_*` — add `Templates` to `GameService` constructor
- `TestAddEmpire_ColonyFromTemplate` — verifies colony is built from template

---

### Task 2: Update `AddEmpire` — create farm group

**Subsystem:** `app/game_service`
**Files:**
- `backend/internal/app/game_service.go`
- `backend/internal/app/game_service_test.go`
**Depends on:** Task 1

**What to do:**

After the colony struct is constructed and before it is appended to
`cluster.Colonies`, add farm group creation.

**1. Add farm group logic to `AddEmpire` in `game_service.go`:**

Insert after the colony struct is built (after the `copy(inventory, ...)` block):

```go
// Build farm group from assembled Farm units in the colony inventory.
// Each colony has at most one FarmGroup; sub-groups are by tech level.
var farmUnits []domain.GroupUnit
for _, inv := range colony.Inventory {
    if inv.Unit == domain.Farm && inv.QuantityAssembled > 0 {
        farmUnits = append(farmUnits, domain.GroupUnit{
            TechLevel: inv.TechLevel,
            Quantity:  inv.QuantityAssembled,
        })
    }
}
if len(farmUnits) > 0 {
    colony.FarmGroups = []domain.FarmGroup{
        {ID: 1, Units: farmUnits},
    }
}
```

`farmUnits` should be sorted by `TechLevel` ascending before creating the
group. Use `slices.SortFunc` from `"slices"` (Go stdlib):

```go
slices.SortFunc(farmUnits, func(a, b domain.GroupUnit) int {
    return int(a.TechLevel) - int(b.TechLevel)
})
```

**2. Add `TestAddEmpire_FarmGroup` to `game_service_test.go`:**

- Use a template with Farm units at two tech levels:
  ```go
  tmpl := domain.ColonyTemplate{
      Kind:      domain.OpenAir,
      TechLevel: 1,
      Inventory: []domain.Inventory{
          {Unit: domain.Farm, TechLevel: 1, QuantityAssembled: 10},
          {Unit: domain.Farm, TechLevel: 2, QuantityAssembled: 5},
          {Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 8},
      },
  }
  ```
- Call `AddEmpire`.
- Assert `len(colony.FarmGroups) == 1`.
- Assert `colony.FarmGroups[0].ID == 1`.
- Assert `len(colony.FarmGroups[0].Units) == 2`.
- Assert sub-groups are sorted by TechLevel: `{TL:1, Qty:10}` then `{TL:2, Qty:5}`.

**3. Add `TestAddEmpire_FarmGroup_NoFarms`:**

- Use a template with no `Farm` units.
- Assert `colony.FarmGroups` is nil (or empty).

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `AddEmpire` creates exactly one `FarmGroup` when Farm units are present
- [ ] `FarmGroup.ID == 1`
- [ ] Sub-groups are one entry per tech level, sorted ascending
- [ ] No `FarmGroup` is created when inventory has no `Farm` units
- [ ] `colony.Inventory` is not modified by farm group creation
- [ ] `TestAddEmpire_FarmGroup` passes
- [ ] `TestAddEmpire_FarmGroup_NoFarms` passes

**Tests to add/update:**
- `TestAddEmpire_FarmGroup` — verifies farm group with multiple TLs
- `TestAddEmpire_FarmGroup_NoFarms` — verifies no group when no Farm units

---

### Task 3: Update `AddEmpire` — create mining groups

**Subsystem:** `app/game_service`
**Files:**
- `backend/internal/app/game_service.go`
- `backend/internal/app/game_service_test.go`
**Depends on:** Task 2

**What to do:**

After farm group creation and before appending the colony to
`cluster.Colonies`, add mining group creation. See the "Mining group
algorithm" section in the Context above for the full specification.

**1. Add a helper function `buildMiningGroups` to `game_service.go`:**

Define as an unexported package-level function (not a method) so it can be
tested independently if needed:

```go
// buildMiningGroups creates one MiningGroup per deposit, distributing
// assembled Mine units as evenly as possible. Remainder units are
// assigned round-robin. Sub-groups within each group are by tech level.
// The returned slice is nil if depositIDs is empty or there are no Mine units.
//
// Note: this is intentionally the simplest assignment that could possibly
// work. Future sprints will rework the algorithm.
func buildMiningGroups(inventory []domain.Inventory, depositIDs []domain.DepositID) []domain.MiningGroup {
    // ... implementation ...
}
```

Steps inside `buildMiningGroups`:

1. Build a mine pool (working copy, sorted by TechLevel ascending):
   ```go
   type mineEntry struct{ TechLevel domain.TechLevel; Quantity int }
   var pool []mineEntry
   for _, inv := range inventory {
       if inv.Unit == domain.Mine && inv.QuantityAssembled > 0 {
           pool = append(pool, mineEntry{inv.TechLevel, inv.QuantityAssembled})
       }
   }
   // sort by TechLevel ascending
   ```

2. If `len(depositIDs) == 0` or `len(pool) == 0`, return nil.

3. Compute `total`, `base`, `remainder` as described in Context.

4. For each `i, depositID := range depositIDs`:
   - `groupQty = base + 1` if `i < remainder`, else `base`.
   - Consume `groupQty` from pool greedily, building `[]GroupUnit`.
   - Append `MiningGroup{ID: MiningGroupID(i+1), Deposit: depositID, Units: units}`.

5. Return the groups.

**2. Call `buildMiningGroups` in `AddEmpire`:**

After the farm group block, find the homeworld planet's deposit IDs and
call `buildMiningGroups`:

```go
// Build mining groups — one per deposit, Mine units split evenly.
var hwDepositIDs []domain.DepositID
for _, p := range cluster.Planets {
    if p.ID == homeWorldID {
        hwDepositIDs = p.Deposits
        break
    }
}
colony.MiningGroups = buildMiningGroups(colony.Inventory, hwDepositIDs)
```

**3. Add tests to `game_service_test.go`:**

**`TestBuildMiningGroups`** — unit test for the helper function directly
(table-driven):

| Case | Mine inventory | Deposits | Expected groups |
|---|---|---|---|
| even split | 20 mines TL1, 3 deposits | IDs 1,2,3 | groups: 7,7,6 mines each |
| remainder round-robin | 10 mines TL1, 3 deposits | IDs 1,2,3 | groups: 4,3,3 mines each |
| multi-TL greedy | 5 TL1 + 5 TL2 mines, 2 deposits | IDs 1,2 | group 1: {TL1:5}, group 2: {TL2:5} |
| no deposits | 10 mines TL1, 0 deposits | none | nil |
| no mines | 0 mines, 3 deposits | IDs 1,2,3 | nil |

**`TestAddEmpire_MiningGroups`** — integration test through `AddEmpire`:

- Use `makeTestClusterWithDeposits(hwPlanetID, []domain.DepositID{10, 11, 12})`
  so the planet has 3 deposits.
- Use a template with `Mine TL1 QuantityAssembled: 9`.
- Call `AddEmpire`.
- Assert `len(colony.MiningGroups) == 3`.
- Assert each group has `Deposit` set to 10, 11, 12 respectively.
- Assert each group has `Units == [{TL:1, Qty:3}]` (9 / 3 = 3 each, no remainder).

**`TestAddEmpire_MiningGroups_NoDeposits`**:

- Use `makeTestCluster(hwPlanetID)` (no deposits).
- Assert `colony.MiningGroups` is nil.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `buildMiningGroups` is an unexported package-level function in `game_service.go`
- [ ] One `MiningGroup` is created per deposit; IDs run 1..N
- [ ] Mine units are split evenly; remainder assigned round-robin from index 0
- [ ] Sub-groups within each group are by tech level, greedy TL-ascending
- [ ] `colony.Inventory` is not modified
- [ ] No mining groups created when planet has no deposits or inventory has no `Mine` units
- [ ] `TestBuildMiningGroups` (table-driven) passes
- [ ] `TestAddEmpire_MiningGroups` passes
- [ ] `TestAddEmpire_MiningGroups_NoDeposits` passes

**Tests to add/update:**
- `TestBuildMiningGroups` — table-driven unit test for `buildMiningGroups`
- `TestAddEmpire_MiningGroups` — integration: 3 deposits, 9 mines → 3 groups of 3
- `TestAddEmpire_MiningGroups_NoDeposits` — no deposits → nil groups

---

### Task 4: Audit, build, and full test suite

**Subsystem:** all
**Files:** all files touched in Tasks 1–3
**Depends on:** Tasks 1–3

**What to do:**

1. **SOUSA import audit:**
   ```bash
   grep -r '"github.com/mdhender/ec/internal/infra\|delivery\|runtime"' backend/internal/app/
   grep -r '"github.com/mdhender/ec/internal/infra"' backend/internal/delivery/
   ```
   Fix any violations.

2. **Inventory not mutated** — verify that neither `buildMiningGroups` nor
   the farm group logic modifies `colony.Inventory`. The inventory is the
   source of truth for what units exist; groups are assignment metadata.

3. **FactoryGroups nil** — verify `colony.FactoryGroups` is never set in
   `AddEmpire` (player assigns via setup orders in a future sprint).

4. **Run `go vet`:**
   ```bash
   cd backend && go vet ./...
   ```

5. **Run full build and test suite:**
   ```bash
   cd backend && go build ./...
   cd backend && go test ./...
   ```

**Acceptance criteria:**
- [ ] No SOUSA import violations
- [ ] `colony.Inventory` is not modified during group creation
- [ ] `colony.FactoryGroups` is nil after `AddEmpire`
- [ ] `cd backend && go vet ./...` passes
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes

**Tests to add/update:**
- None — audit task only.

---

## Task Summary

| Task | Title                                      | Status | Depends On | Agent/Thread | Notes |
|------|--------------------------------------------|--------|------------|--------------|-------|
| 1    | AddEmpire: colony from template            | TODO   | Sprint 8   |              |       |
| 2    | AddEmpire: farm group creation             | TODO   | 1          |              |       |
| 3    | AddEmpire: mining group creation           | TODO   | 2          |              |       |
| 4    | Audit, build, and full test suite          | TODO   | 1–3        |              |       |
