# Sprint 7: Homeworld Placement

**Pass:** Pass 1
**Goal:** Restructure the game file layout around a single data directory and implement homeworld creation, race assignment, and empire placement.
**Predecessor:** Sprint 6

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

This sprint has two concerns: (1) cleaning up a structural mismatch in the
game file layout, and (2) implementing homeworld placement and race creation.

After completing a task, update sprints/sprint-7.md: check off acceptance
criteria (change `[ ]` to `[x]`) and change the task status from TODO to DONE
in the Task Summary table at the bottom of the file.

### File layout (after this sprint)

All game data for one game lives under a single `data-path` directory:

```
data-path/
  cluster.json   ← domain.Cluster (written by `create cluster`)
  game.json      ← domain.Game (empires, races, homeworlds — no cluster data)
  auth.json      ← domain.AuthConfig (magic links)
  1/             ← empire 1 subdirectory (orders, reports)
  2/
  ...
```

`game.json` does **not** embed the cluster. Both files are read together when
the full game state is needed (e.g. `create homeworld`, `create empire`).

### Command workflow (after this sprint)

```
cli create game      --data-path /dir                 → game.json + auth.json
cli create cluster   --data-path /dir [--seed1 N --seed2 N]  → cluster.json
cli create homeworld --data-path /dir [--planet N] [--min-distance N]
cli create empire    --data-path /dir [--empire N] [--name S] [--homeworld N]
cli show magic-link  --data-path /dir --empire N      → unchanged
```

### Architecture constraints

This project enforces the SOUSA layering rule: `domain ← app ← infra/delivery ← runtime`.

- `domain` — pure types and game rules, no framework imports
- `app` — use-case services and port interfaces; no SQL, HTTP, or file I/O
- `infra/filestore` — implements app ports using the filesystem
- `delivery/cli` — thin command wrappers; calls app services, no game logic
- `runtime/cli` — wires concrete types; the only layer that knows about infra

`delivery` must not import `infra`. `runtime` is the only layer that
instantiates and injects concrete implementations.

### Key files

**Before this sprint (will change):**
- `backend/internal/domain/game_config.go` — `GameConfig`, `EmpireEntry` (to be deleted)
- `backend/internal/domain/game.go` — `Game`, `Empire` (to be extended)
- `backend/internal/domain/cluster.go` — `Cluster`, `Coords` (to be extended)
- `backend/internal/app/cluster_ports.go` — `ClusterReader`, `ClusterWriter`, `GameWriter`
- `backend/internal/app/cluster_service.go` — `ClusterService.CreateGame` (to be deleted)
- `backend/internal/app/game_config_ports.go` — `GameConfigStore` (methods change)
- `backend/internal/app/game_config_service.go` — `GameConfigService` (to be updated)
- `backend/internal/infra/filestore/cluster.go` — `WriteCluster`, `WriteGame` (to be updated/deleted)
- `backend/internal/infra/filestore/game_config.go` — game/auth config I/O (to be updated)
- `backend/internal/delivery/cli/cluster.go` — `CmdCreateCluster`, `CmdCreateGameState`
- `backend/internal/delivery/cli/game_config.go` — `CmdCreateGame`, `CmdAddEmpire`, etc.
- `backend/internal/runtime/cli/cli.go` — wiring

### Key types (current)

```go
// domain/game.go
type Game struct { Cluster Cluster; Empires []Empire }
type Empire struct { ID EmpireID; Name string; HomeWorld PlanetID; Colonies []ColonyID; Ships []ShipID }

// domain/game_config.go  ← DELETED in Task 5
type GameConfig struct { Empires []EmpireEntry }
type EmpireEntry struct { Empire int; Active bool }

// domain/cluster.go
type Coords struct { X, Y, Z int }
type Planet struct { ID PlanetID; Kind PlanetKind; Habitability int; Deposits []DepositID }
// PlanetKind: Terrestrial = 1, AsteroidBelt = 2, GasGiant = 3
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

### Task 1: Restructure cluster file location

**Subsystem:** `delivery/cli`, `app/cluster_ports`, `app/cluster_service`, `infra/filestore`
**Files:**
- `backend/internal/delivery/cli/cluster.go`
- `backend/internal/app/cluster_ports.go`
- `backend/internal/app/cluster_service.go`
- `backend/internal/infra/filestore/cluster.go`
**Depends on:** None

**What to do:**

The `create cluster` command currently writes to an arbitrary file path via
`--path`. Change it to write to `<data-path>/cluster.json` via `--data-path`.

1. **`delivery/cli/cluster.go` — `CmdCreateCluster`**: rename flag `--path` to
   `--data-path` (default `"testdata"`). Pass `*dataPath` to
   `svc.CreateCluster`; the service now computes the filename internally.

2. **`app/cluster_ports.go` — `ClusterWriter` interface**: change signature from
   `WriteCluster(path string, cluster domain.Cluster, overwrite bool) error`
   to `WriteCluster(dataPath string, cluster domain.Cluster, overwrite bool) error`.
   The `dataPath` parameter is a directory; implementations write to
   `<dataPath>/cluster.json`.

3. **`app/cluster_service.go` — `ClusterService.CreateCluster`**: change
   `outputPath string` parameter to `dataPath string`. Pass it directly to
   `s.Writer.WriteCluster`.

4. **`infra/filestore/cluster.go` — `WriteCluster`**: construct the output
   path as `filepath.Join(dataPath, "cluster.json")` internally. Remove the
   assumption that the caller provides a file path. Keep the directory-exists
   check on `dataPath` itself (not the parent of a file path).

Do **not** change `ClusterReader` or `ClusterService.CreateGame` in this task;
those are removed in Task 2.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `CmdCreateCluster` flag is `--data-path`, not `--path`
- [ ] Running `cli create cluster --data-path /some/dir` writes `/some/dir/cluster.json`
- [ ] `ClusterWriter.WriteCluster` signature takes a directory path

**Tests to add/update:**
- No new tests — validated via build. The cluster service has no dedicated unit tests.

---

### Task 2: Delete `create game-state` and related code

**Subsystem:** `delivery/cli`, `app/cluster_ports`, `app/cluster_service`, `infra/filestore`, `runtime/cli`
**Files:**
- `backend/internal/delivery/cli/cluster.go`
- `backend/internal/app/cluster_ports.go`
- `backend/internal/app/cluster_service.go`
- `backend/internal/infra/filestore/cluster.go`
- `backend/internal/runtime/cli/cli.go`
**Depends on:** Task 1

**What to do:**

`CmdCreateGameState` is superseded by the new file layout. Remove it and all
code it depends on exclusively.

1. **`delivery/cli/cluster.go`**: delete `CmdCreateGameState` entirely.

2. **`app/cluster_service.go`**: delete `ClusterService.CreateGame`. Remove
   the `GameWriter GameWriter` field from the `ClusterService` struct.

3. **`app/cluster_ports.go`**: delete the `GameWriter` interface. Delete the
   `ClusterReader` interface (its single caller, `CreateGame`, is gone; a new
   `ClusterStore` interface will replace it in Task 6).

4. **`infra/filestore/cluster.go`**: delete `WriteGame`. Delete `ReadCluster`
   if it is no longer needed after removing `ClusterReader`. (If `ReadCluster`
   is referenced by other code, leave it and note the dependency.)

5. **`runtime/cli/cli.go`**: remove `GameWriter: store` from the
   `clusterSvc` literal. Remove `deliverycli.CmdCreateGameState(clusterSvc)`
   from the `createCmd.Subcommands` slice.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `CmdCreateGameState` does not exist anywhere in the codebase
- [ ] `ClusterService` has no `GameWriter` field
- [ ] `GameWriter` interface does not exist
- [ ] `cli create game-state` is not a valid command

**Tests to add/update:**
- None — deleted code had no tests.

---

### Task 3: Domain model — add Empire.Active

**Subsystem:** `domain`
**Files:**
- `backend/internal/domain/game.go`
**Depends on:** None (parallel with Tasks 1–2)

**What to do:**

Add the `Active bool` field to `domain.Empire`. This is an additive change
that does not break any existing code.

1. **`domain/game.go`**: add `Active bool` to `Empire`:

   ```go
   type Empire struct {
       ID        EmpireID
       Name      string
       Active    bool
       HomeWorld PlanetID
       Colonies  []ColonyID
       Ships     []ShipID
   }
   ```

Do **not** delete `GameConfig` or `EmpireEntry` in this task — they are still
referenced by `app` and `infra` layers. Deletion happens in Task 5 after
storage is migrated.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `domain.Empire` has an `Active bool` field
- [ ] `domain.GameConfig` and `domain.EmpireEntry` still exist (not yet deleted)

**Tests to add/update:**
- None — pure additive model change with no new logic.

---

### Task 4: Domain — add Race/HomeWorld types, remove Cluster from Game, add Coords.Distance

**Subsystem:** `domain`
**Files:**
- `backend/internal/domain/game.go`
- `backend/internal/domain/cluster.go`
**Depends on:** Tasks 2, 3

**What to do:**

Add the domain types needed for homeworld placement and race tracking. Also
remove `Cluster` from `Game` (it now lives in a separate `cluster.json`).

**Important:** Task 2 must be completed first — it deletes
`ClusterService.CreateGame`, which references `domain.Game{Cluster: cluster}`.
Without that deletion, removing the `Cluster` field from `Game` would break
the build.

1. **`domain/game.go`** — add:

   ```go
   type RaceID int  // equals the PlanetID of the race's homeworld

   type Race struct {
       ID        RaceID
       HomeWorld PlanetID
       Empires   []EmpireID // maximum 25
   }
   ```

   Update `domain.Game`:
   ```go
   type Game struct {
       Empires           []Empire
       Races             []Race
       ActiveHomeWorldID PlanetID // 0 means none set
   }
   ```
   Note: `Cluster` is **removed** from `Game` — it lives in `cluster.json`
   separately.

   Update `domain.Empire`:
   ```go
   type Empire struct {
       ID        EmpireID
       Name      string
       Active    bool      // added in Task 3
       Race      RaceID    // new
       HomeWorld PlanetID
       Colonies  []ColonyID
       Ships     []ShipID
   }
   ```

2. **`domain/cluster.go`** — add a `Distance` method to `Coords`:

   ```go
   import "math"

   func (c Coords) Distance(other Coords) float64 {
       dx := float64(c.X - other.X)
       dy := float64(c.Y - other.Y)
       dz := float64(c.Z - other.Z)
       return math.Sqrt(dx*dx + dy*dy + dz*dz)
   }
   ```

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `domain.Race`, `domain.RaceID` exist
- [ ] `domain.Game` has `Races []Race` and `ActiveHomeWorldID PlanetID`; no `Cluster` field
- [ ] `domain.Empire` has `Race RaceID`
- [ ] `domain.Coords` has a `Distance(Coords) float64` method
- [ ] Field spelling is consistently `HomeWorld` (capital W) everywhere

**Tests to add/update:**
- `TestCoordsDistance` in `backend/internal/domain/cluster_test.go` (new file
  if it does not exist) — table-driven: known coordinate pairs and expected
  distances (include axis-aligned, diagonal, and zero-distance cases).

---

### Task 5: Migrate game config storage to domain.Game; delete GameConfig

**Subsystem:** `app/game_config_ports`, `app/game_config_service`, `infra/filestore`
**Files:**
- `backend/internal/app/game_config_ports.go`
- `backend/internal/app/game_config_service.go`
- `backend/internal/infra/filestore/game_config.go`
- `backend/internal/app/game_config_service_test.go`
- `backend/internal/infra/filestore/game_config_test.go`
- `backend/internal/domain/game_config.go`
**Depends on:** Task 4 (domain.Game must have its final shape first)

**What to do:**

Replace all references to `domain.GameConfig` with `domain.Game` throughout
the game config storage layer, then delete `GameConfig` and `EmpireEntry`.

1. **`app/game_config_ports.go` — `GameConfigStore` interface**:
   - Rename `GameConfigExists` → `GameExists`
   - Rename `ReadGameConfig(path string) (domain.GameConfig, error)` →
     `ReadGame(path string) (domain.Game, error)`
   - Rename `WriteGameConfig(path string, cfg domain.GameConfig) error` →
     `WriteGame(path string, game domain.Game) error`
   - Keep unchanged: `ValidateDir`, `AuthConfigExists`, `ReadAuthConfig`,
     `WriteAuthConfig`, `CreateEmpireDir`

2. **`app/game_config_service.go`**: update all service methods to read/write
   `domain.Game` via the renamed interface methods.
   - `CreateGame(dirPath)`: write `domain.Game{Empires: []domain.Empire{}}` via
     `WriteGame`; write empty `domain.AuthConfig` via `WriteAuthConfig`.
   - `AddEmpire(dirPath, empireNo)`: read `domain.Game`, append a new
     `domain.Empire{ID: domain.EmpireID(empireNo), Active: true}`, write back.
   - `RemoveEmpire(dirPath, empireNo)`: read `domain.Game`, set
     `Empires[i].Active = false`, write back.
   - `ShowMagicLink`: unchanged (reads `auth.json` only).

3. **`infra/filestore/game_config.go`**: implement the renamed methods.
   - `GameExists`: check for `game.json` (same file, new method name).
   - `ReadGame`: unmarshal `game.json` into `domain.Game`.
   - `WriteGame`: marshal `domain.Game` and write to `game.json`.
   - Keep all `AuthConfig` methods unchanged.

4. **`app/game_config_service_test.go`**: rewrite `mockGameConfigStore` to
   implement the new interface (store `domain.Game` instead of
   `domain.GameConfig`). Update all test cases to match. All existing test
   scenarios must be preserved.

5. **`infra/filestore/game_config_test.go`**: update `TestGameConfigRoundTrip`
   to use `domain.Game` instead of `domain.GameConfig`. Rename to
   `TestGameRoundTrip`. Update assertions to match the new type.

6. **`domain/game_config.go`**: delete `GameConfig` and `EmpireEntry`.
   Keep `AuthConfig` and `AuthLink` — they are still used by the auth layer.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes (all existing test cases still present and passing)
- [ ] No references to `domain.GameConfig` or `domain.EmpireEntry` remain anywhere
- [ ] `GameConfigStore` interface has `GameExists`, `ReadGame`, `WriteGame`
- [ ] `game.json` on disk now marshals as `domain.Game`

**Tests to add/update:**
- `TestCreateGame`, `TestAddEmpire`, `TestRemoveEmpire` in
  `backend/internal/app/game_config_service_test.go` — rewrite mock to use
  `domain.Game`; all existing scenarios must be preserved with the new types.
- `TestGameRoundTrip` in `backend/internal/infra/filestore/game_config_test.go`
  — update to use `domain.Game`.

---

### Task 6: Add ClusterStore interface and update filestore.ReadCluster

**Subsystem:** `app/cluster_ports`, `infra/filestore`
**Files:**
- `backend/internal/app/cluster_ports.go`
- `backend/internal/infra/filestore/cluster.go`
**Depends on:** Task 1 (WriteCluster already uses dataPath)

**What to do:**

Add the `ClusterStore` interface and update `filestore.ReadCluster` to accept
a directory path instead of a file path.

1. **`app/cluster_ports.go`** — add:

   ```go
   // ClusterStore reads and writes cluster.json from the data directory.
   type ClusterStore interface {
       ReadCluster(dataPath string) (domain.Cluster, error)
       WriteCluster(dataPath string, cluster domain.Cluster, overwrite bool) error
   }
   ```

   `ClusterWriter` (from Task 1) already has the correct `WriteCluster`
   signature. `ClusterStore` unifies read and write under one interface.

2. **`infra/filestore/cluster.go`** — update `ReadCluster`:
   - Change from reading the path directly to constructing
     `filepath.Join(dataPath, "cluster.json")` internally.
   - The parameter name changes from `path` to `dataPath`.

After this task, `filestore.Store` satisfies both `ClusterWriter` and the new
`ClusterStore` interface.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `ClusterStore` interface exists in `app/cluster_ports.go`
- [ ] `filestore.ReadCluster` takes a directory path and reads `<dataPath>/cluster.json`
- [ ] `filestore.Store` satisfies `ClusterStore`

**Tests to add/update:**
- None — `ReadCluster` had no tests and the change is mechanical.

---

### Task 7: Implement CreateHomeWorld service logic

**Subsystem:** `app/game_config_service`
**Files:**
- `backend/internal/app/game_config_ports.go`
- `backend/internal/app/game_config_service.go`
- `backend/internal/app/game_config_service_test.go`
**Depends on:** Tasks 5, 6

**What to do:**

Implement the `CreateHomeWorld` use case. Creating a homeworld selects a
terrestrial planet from the cluster, sets its habitability to 25, creates a
Race (with `ID == planet.ID`), appends it to `game.Races`, and sets
`game.ActiveHomeWorldID` to that planet's ID.

**1. Update `GameConfigService` struct in `app/game_config_service.go`:**

Add a `Cluster` field:

```go
type GameConfigService struct {
    Store   GameConfigStore
    Cluster ClusterStore    // new — reads/writes cluster.json
}
```

**2. Add `GameConfigService.CreateHomeWorld`:**

```go
func (s *GameConfigService) CreateHomeWorld(dataPath string, planetID domain.PlanetID, minDistance int) (domain.PlanetID, error)
```

Behavior:

- Read `game.json` via `s.Store.ReadGame(dataPath)`.
- Read `cluster.json` via `s.Cluster.ReadCluster(dataPath)`.
- If `planetID != 0` (GM specified `--planet`): locate the planet in the
  cluster by ID (walk `cluster.Planets`, match by `Planet.ID`). If not found,
  return error. If not `Terrestrial`, return error. If already a homeworld
  (its ID appears in any `game.Races[i].HomeWorld`), return error. Skip the
  min-distance check entirely.
- If `planetID == 0` (auto-select): collect all planets where
  `Kind == Terrestrial` and whose ID does not appear in any
  `game.Races[i].HomeWorld`. For each candidate, find its system coordinates
  (walk `cluster.Stars` to find the star whose `Orbits` array contains the
  planet ID, then look up the system by matching `star.System` against
  `cluster.Systems[i].ID` to get `cluster.Systems[i].Location`).
  **Important:** look up systems by ID, not by slice index.
  Filter out candidates whose system is strictly within `minDistance` units of
  any existing homeworld's system (reject when
  `distance < float64(minDistance)`). Use `Coords.Distance`. If no candidates
  remain, return a hard error: `"no terrestrial planets available at
  min-distance %d"`. Pick one at random using `math/rand/v2`.
- Set `planet.Habitability = 25` in the cluster (find the planet in
  `cluster.Planets` by ID and update it).
- Append `Race{ID: domain.RaceID(planetID), HomeWorld: planetID, Empires: nil}`
  to `game.Races`.
- Set `game.ActiveHomeWorldID = planetID`.
- Write updated `game.json` via `s.Store.WriteGame(dataPath, game)`.
- Write updated `cluster.json` via `s.Cluster.WriteCluster(dataPath, cluster, true)`.
- Return `(planetID, nil)`.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `GameConfigService` has a `Cluster ClusterStore` field
- [ ] `CreateHomeWorld` auto-selects a terrestrial planet when `planetID == 0`
- [ ] `CreateHomeWorld` uses `--planet N` when `planetID != 0`; skips min-distance
- [ ] Habitability is set to 25 on the chosen planet
- [ ] A race is created with `ID == PlanetID`
- [ ] `game.ActiveHomeWorldID` is set after creation
- [ ] Fails with clear error if planet not found
- [ ] Fails with clear error if planet is not terrestrial
- [ ] Fails with clear error if planet is already a homeworld
- [ ] Fails with clear error if no terrestrial planets available
- [ ] Distance filtering rejects candidates where `distance < float64(minDistance)`

**Tests to add/update:**
- `TestCreateHomeWorld_AutoSelect` in `backend/internal/app/game_config_service_test.go`
  — mock cluster with several terrestrials; verify a race is created and
  `ActiveHomeWorldID` is set.
- `TestCreateHomeWorld_PlanetFlag` — specify `--planet` for a valid terrestrial;
  verify it succeeds and skips distance check.
- `TestCreateHomeWorld_PlanetFlagSkipsDistance` — specify `--planet` for a
  planet that would fail min-distance; verify it succeeds.
- `TestCreateHomeWorld_PlanetNotFound` — `--planet` for a planet ID that does
  not exist in the cluster; verify error.
- `TestCreateHomeWorld_NotTerrestrial` — `--planet` for a gas giant; verify error.
- `TestCreateHomeWorld_AlreadyHomeWorld` — `--planet` for a planet that is
  already a homeworld; verify error.
- `TestCreateHomeWorld_MinDistance` — two existing homeworlds; verify a
  candidate too close is rejected; verify a distant one is accepted.
- `TestCreateHomeWorld_NoTerrestrials` — cluster with no terrestrial planets;
  verify hard error.

---

### Task 8: Add `create homeworld` CLI command and runtime wiring

**Subsystem:** `delivery/cli`, `runtime/cli`
**Files:**
- `backend/internal/delivery/cli/game_config.go`
- `backend/internal/runtime/cli/cli.go`
**Depends on:** Task 7

**What to do:**

Wire the `CreateHomeWorld` service method into a CLI command.

**1. Add `CmdCreateHomeWorld` to `delivery/cli/game_config.go`:**

```
Name:  "homeworld"
Usage: "cli create homeworld [FLAGS]"
Flags: --data-path (required), --planet (int, default 0 = auto), --min-distance (int, default 3)
Exec:  validate --data-path not empty; call svc.CreateHomeWorld; print "homeworld created: planet <id>"
```

The service returns `(domain.PlanetID, error)`, so the CLI can print the
planet ID.

**2. Update `runtime/cli/cli.go`:**

- Add `Cluster: store` to the `gameConfigSvc` literal.
- Add `deliverycli.CmdCreateHomeWorld(gameConfigSvc)` to `createCmd.Subcommands`.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cli create homeworld --data-path /dir` picks a terrestrial planet and prints its ID
- [ ] `cli create homeworld --data-path /dir --planet N` uses planet N
- [ ] `cli create homeworld --data-path /dir --min-distance 5` enforces 5-unit separation
- [ ] Missing `--data-path` produces a clear error

**Tests to add/update:**
- None — CLI commands are validated via build and integration testing in Task 12.

---

### Task 9: Update AddEmpire service to assign race and starting colony

**Subsystem:** `app/game_config_service`
**Files:**
- `backend/internal/app/game_config_service.go`
- `backend/internal/app/game_config_service_test.go`
**Depends on:** Task 7

**What to do:**

Update `AddEmpire` to require an active homeworld, assign the empire to
that homeworld's race, create a starting colony on the homeworld planet, and
accept an empire name (scrubbed before storing).

**1. Update `GameConfigService.AddEmpire` signature:**

```go
func (s *GameConfigService) AddEmpire(dataPath string, empireNo int, name string, homeWorldID domain.PlanetID) (int, string, error)
```

Behavior:

- Read `game.json` and `cluster.json`.
- **HomeWorld resolution**: if `homeWorldID == 0`, use `game.ActiveHomeWorldID`.
  If still 0, return error: `"no active homeworld; use --homeworld or run create homeworld first"`.
  If `homeWorldID != 0`, look it up in `game.Races` by `Race.HomeWorld == homeWorldID`.
  If not found, return error: `"homeworld <N> does not exist"`.
- **Limit**: if `len(race.Empires) >= 25`, return error: `"homeworld <N> is full (25 empires)"`.
- **Name scrubbing**: call `scrubEmpireName(name)` (see below). If the result
  is empty, return error: `"empire name is required"`.
- **Empire number**: auto-assign (max existing ID + 1) if `empireNo == 0`.
  If `empireNo` is already taken, return error as before.
- **Starting colony**: find the homeworld planet in the cluster (look up by
  `Planet.ID`, not by slice index). Find its system (walk `cluster.Stars` to
  find which star has the planet in its `Orbits`, then look up the system by
  matching `star.System` against `cluster.Systems[i].ID`). Create:
  ```go
  colony := domain.Colony{
      ID:        domain.ColonyID(len(cluster.Colonies) + 1),
      Empire:    domain.EmpireID(empireNo),
      Location:  system.Location,
      TechLevel: 1,
  }
  ```
  Append `colony` to `cluster.Colonies`.
- Create:
  ```go
  empire := domain.Empire{
      ID:        domain.EmpireID(empireNo),
      Name:      scrubbedName,
      Active:    true,
      Race:      race.ID,
      HomeWorld: homeWorldID,
      Colonies:  []domain.ColonyID{colony.ID},
  }
  ```
- Append empire to `game.Empires`. Append `domain.EmpireID(empireNo)` to
  `race.Empires`.
- Generate magic link UUID; update `auth.json` as before.
- Write `game.json` and `cluster.json` back.
- Call `s.Store.CreateEmpireDir(dataPath, empireNo)` as before.
- Return `(empireNo, uuid, nil)`.

**`scrubEmpireName(name string) string`** — unexported function in
`app/game_config_service.go`:

- Strip any character that is not a Unicode letter, digit, space, hyphen `-`,
  underscore `_`, period `.`, or comma `,`. This removes HTML-special chars
  (`<`, `>`, `&`, `"`, `'`) and shell-special chars (`` ` ``, `$`, `;`, `|`,
  `(`, `)`, `{`, `}`, `[`, `]`, `\`, `*`, `?`, `!`, `#`, `^`, `~`, newlines).
- Compress runs of spaces to a single space.
- Trim leading and trailing spaces.

**Preserving existing behaviors:** All existing `AddEmpire` behaviors must
continue to work:
- Auto-assign next empire number when `empireNo == 0`
- Reject duplicate empire numbers
- Generate and store magic link UUID
- Create empire data directory

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go test ./...` passes
- [ ] `AddEmpire` fails if no active homeworld and no `homeWorldID` given
- [ ] `AddEmpire` with explicit `homeWorldID` fails if homeworld does not exist
- [ ] A 26th empire on the same homeworld is rejected
- [ ] Empire is added to `race.Empires` in `game.json`
- [ ] Empire `Race` and `HomeWorld` fields are set
- [ ] A starting colony is created in `cluster.Colonies` and in `empire.Colonies`
- [ ] Empire name is scrubbed: HTML/shell-special chars removed, spaces compressed, trimmed
- [ ] Scrubbing a name to empty string returns an error
- [ ] Existing behaviors preserved: auto-assign, duplicate rejection, magic link, empire dir

**Tests to add/update:**
- `TestAddEmpire_RequiresHomeWorld` — call `AddEmpire` with no active homeworld;
  verify error.
- `TestAddEmpire_HomeWorldOverride` — call with explicit `homeWorldID` that
  exists; verify empire is assigned to that race.
- `TestAddEmpire_HomeWorldNotFound` — call with `homeWorldID` that does not
  exist in `game.Races`; verify error.
- `TestAddEmpire_HomeWorldFull` — race already has 25 empires; verify error.
- `TestScrubEmpireName` — table-driven: inputs with HTML chars, shell chars,
  extra spaces, leading/trailing blanks, all-special input (→ empty), normal
  input (→ unchanged).
- Update existing `TestAddEmpire` subtests to pass the new parameters (`name`
  and `homeWorldID`).

---

### Task 10: Update `create empire` CLI command with new flags

**Subsystem:** `delivery/cli`
**Files:**
- `backend/internal/delivery/cli/game_config.go`
**Depends on:** Task 9

**What to do:**

Update `CmdAddEmpire` to pass the new `name` and `homeWorldID` parameters to
the updated `AddEmpire` service method.

1. **`delivery/cli/game_config.go` — `CmdAddEmpire`**: add flags:
   - `--name` (string, default `""`) — empire name; required after scrubbing.
   - `--homeworld` (int, default `0`) — override homeworld planet ID.

   Pass both to `svc.AddEmpire(dataPath, empireNo, name, homeWorldID)`. Update
   the success print to include the empire name:
   ```
   added empire %d (%s), magic link: %s
   ```

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds
- [ ] `CmdAddEmpire` has `--name` and `--homeworld` flags
- [ ] `cli create empire --data-path /dir --name "Test"` passes the name to the service
- [ ] `cli create empire --data-path /dir --homeworld 42` passes the homeworld ID

**Tests to add/update:**
- None — CLI commands are validated via build and integration testing in Task 12.

---

### Task 11: Audit for code smells and SOUSA compliance

**Subsystem:** all
**Files:** all files touched in Tasks 1–10
**Depends on:** Tasks 1–10

**What to do:**

Review all code changed in this sprint for SOUSA layering violations, code
smells, and consistency issues. The canonical SOUSA reference is
`docs/SOUSA.md`.

1. **SOUSA import audit** — verify no layering violations were introduced:
   ```bash
   # domain must not import app, infra, delivery, or runtime
   grep -r '"github.com/mdhender/ec/internal/app\|infra\|delivery\|runtime"' backend/internal/domain/

   # app must not import infra, delivery, runtime, or framework packages
   grep -r '"github.com/mdhender/ec/internal/infra\|delivery\|runtime"' backend/internal/app/
   grep -r '"github.com/peterbourgon/ff\|labstack/echo"' backend/internal/app/

   # delivery must not import infra
   grep -r '"github.com/mdhender/ec/internal/infra"' backend/internal/delivery/
   ```
   Fix any violations found.

2. **Game logic in wrong layer** — verify that `delivery/cli` commands contain
   no game logic (homeworld selection, distance calculations, name scrubbing).
   All game logic must live in `app` or `domain`.

3. **Unused code** — search for any dead code left behind by the refactoring:
   - Unused imports
   - Unreferenced types or functions
   - Orphaned test helpers

4. **Naming consistency** — verify:
   - `HomeWorld` (capital W) is used consistently in all field names
   - No stale `GameConfig`/`EmpireEntry` references remain
   - Port interface names match their implementations

5. **Error quality** — spot-check that error messages from new service methods
   include enough context to diagnose problems (which planet, which homeworld,
   what distance).

6. **Run `go vet`:**
   ```bash
   cd backend && go vet ./...
   ```
   Fix any warnings.

**Acceptance criteria:**
- [ ] No SOUSA import violations exist
- [ ] No game logic in `delivery/cli` — all logic is in `app` or `domain`
- [ ] No unused code (imports, types, functions, test helpers)
- [ ] `HomeWorld` spelling is consistent everywhere
- [ ] No stale `GameConfig`/`EmpireEntry` references
- [ ] `cd backend && go vet ./...` passes
- [ ] All error messages include sufficient diagnostic context

**Tests to add/update:**
- None — this is an audit task. Fix any issues found in the files where they
  occur.

---

### Task 12: Verify build, clean up, run full test suite

**Subsystem:** all
**Files:** any remaining references
**Depends on:** Task 11

**What to do:**

1. Run `cd backend && go build ./...` — fix any remaining compilation errors.
2. Run `cd backend && go test ./...` — fix any test failures.
3. Run `cd backend && go vet ./...` — fix any vet warnings.
4. Search for stale references:
   ```bash
   grep -r "GameConfig\|EmpireEntry\|GameWriter\|CmdCreateGameState\|ClusterService\.CreateGame\|ClusterReader" backend/
   ```
   Remove or fix any found outside of comments or test names that reference
   the old concepts.
   **Note:** `GameConfigService.CreateGame` is still valid — do not remove it.
5. Run `go mod tidy` to clean unused transitive dependencies.
6. Smoke-test the CLI sequence end-to-end using a temp directory:
   ```bash
   mkdir /tmp/ec-test
   cli create game --data-path /tmp/ec-test
   cli create cluster --data-path /tmp/ec-test
   cli create homeworld --data-path /tmp/ec-test
   cli create empire --data-path /tmp/ec-test --name "Test Empire"
   cli show magic-link --data-path /tmp/ec-test --empire 1
   ```
   Each command must succeed and the expected files must exist.

**Acceptance criteria:**
- [ ] `cd backend && go build ./...` succeeds with zero errors
- [ ] `cd backend && go test ./...` passes all tests
- [ ] `cd backend && go vet ./...` passes
- [ ] No source files reference `domain.GameConfig`, `domain.EmpireEntry`, `CmdCreateGameState`, `GameWriter`, or `ClusterReader`
- [ ] End-to-end smoke test runs without errors

**Tests to add/update:**
- None beyond what earlier tasks specify.

---

## Task Summary

| Task | Title                                                  | Status | Depends On | Agent/Thread | Notes |
|------|--------------------------------------------------------|--------|------------|--------------|-------|
| 1    | Restructure cluster file location                      | TODO   | —          |              |       |
| 2    | Delete `create game-state` and related code            | TODO   | 1          |              |       |
| 3    | Domain: add Empire.Active                              | TODO   | —          |              |       |
| 4    | Domain: add Race/HomeWorld types, Coords.Distance      | TODO   | 2, 3       |              |       |
| 5    | Migrate game config storage; delete GameConfig         | TODO   | 4          |              |       |
| 6    | Add ClusterStore interface + update ReadCluster        | TODO   | 1          |              |       |
| 7    | Implement CreateHomeWorld service logic                 | TODO   | 5, 6       |              |       |
| 8    | Add `create homeworld` CLI command + wiring            | TODO   | 7          |              |       |
| 9    | Update AddEmpire service with race + colony            | TODO   | 7          |              |       |
| 10   | Update `create empire` CLI with new flags              | TODO   | 9          |              |       |
| 11   | Audit for code smells and SOUSA compliance             | TODO   | 1–10       |              |       |
| 12   | Verify build, clean up, full test suite                | TODO   | 11         |              |       |
