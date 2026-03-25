# Sprint 4: Cluster Generator

**Pass:** Pass 1
**Goal:** Add a deterministic cluster generator to the CLI that produces normalized domain objects, with SOUSA-compliant layering.
**Predecessor:** Sprint 3

{{< callout type="warning" >}}
This sprint document was written retroactively after implementation.
The generator was built first, then reviewed for SOUSA compliance.
The review found six violations and six code smells, all fixed before closing.
{{< /callout >}}

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

This sprint adds cluster generation to the CLI. A cluster is the star map every game starts from: 100 stars placed in a 31×31×31 cube, each with orbits containing terrestrial planets, asteroid belts, and gas giants. Planets have habitability ratings and natural resource deposits (gold, fuel, metallics, non-metallics).

Generation is deterministic — the same PRNG seeds produce the same cluster. The generator uses internal tree-shaped types for convenience during generation, then normalizes the result into flat `domain.Cluster` with shuffled IDs before returning.

The project follows SOUSA layering (see `docs/SOUSA.md`). Dependencies flow inward only: `domain ← app ← infra / delivery ← runtime`.

**Key files:**
- `backend/internal/domain/clustergen/generate.go` — generation logic (unexported tree types, exported `GenerateCluster`)
- `backend/internal/domain/clustergen/stats.go` — `ClusterStats` aggregation over `domain.Cluster`
- `backend/internal/domain/cluster.go` — `domain.Cluster`, `System`, `Star`, `Planet`, `Deposit` types
- `backend/internal/app/cluster_ports.go` — `ClusterReader`, `ClusterWriter`, `GameWriter` port interfaces
- `backend/internal/app/cluster_service.go` — `ClusterService` with `CreateCluster`, `TestCluster`, `CreateGame`
- `backend/internal/infra/filestore/cluster.go` — JSON read/write for clusters and games
- `backend/internal/delivery/cli/cluster.go` — thin cobra commands
- `backend/internal/delivery/cli/report.go` — report formatting (`WriteClusterReport`, `WriteStatsReport`)
- `backend/internal/runtime/cli/cli.go` — wires infra → app → delivery for CLI

**Key types/functions:**
- `clustergen.GenerateCluster(r *prng.Rand) (domain.Cluster, error)` — the public API
- `clustergen.ClusterStats` / `ClusterStats.Collect(domain.Cluster)` — stats aggregation
- `app.ClusterService` — orchestrates use cases
- `filestore.Store.ReadCluster`, `WriteCluster`, `WriteGame` — persistence

**Build/test commands:**
```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/cli/
```

**Constraints reminder:**
- `domain` must not import `app`, `infra`, `delivery`, or `runtime`
- `app` must not import Echo, SQLite, Cobra, or filesystem adapters
- `delivery` must not import `infra` — they are peers
- Generator tree types must be unexported — only `GenerateCluster` is public

---

## Tasks

### Task 1: Domain types for clusters

**Subsystem:** `domain`
**Files:** `backend/internal/domain/cluster.go`, `backend/internal/domain/game.go`
**Depends on:** None

**What to do:**
Define the normalized domain types for the cluster. These are flat, ID-referenced structures — not a tree.

Types: `SystemID`, `System`, `Coords`, `StarID`, `Star`, `PlanetID`, `Planet`, `PlanetKind` (Terrestrial, AsteroidBelt, GasGiant), `DepositID`, `Deposit`, `NaturalResource` (GOLD, FUEL, METALLICS, NONMETALLICS), `Cluster` (containing slices of Systems, Stars, Planets, Deposits, Colonies, Ships).

`Game` wraps a `Cluster` and a slice of `Empire`.

**Acceptance criteria:**
- [x] `go build ./internal/domain/...` succeeds
- [x] `Coords.Less` provides lexicographic ordering (X → Y → Z)
- [x] `PlanetKind` and `NaturalResource` have `String()` methods
- [x] `Star.Orbits` is a `[10]PlanetID` fixed-size array
- [x] No imports of outer layers

**Tests to add/update:**
- None — pure type definitions

---

### Task 2: Cluster generation logic in domain/clustergen

**Subsystem:** `domain/clustergen`
**Files:** `backend/internal/domain/clustergen/generate.go`
**Depends on:** Task 1

**What to do:**
Implement the cluster generator. The generator uses unexported tree types (`system`, `star`, `planet`, `deposit`) internally for convenience during the generation process.

The single exported function is `GenerateCluster(r *prng.Rand) (domain.Cluster, error)`. It:
1. Generates 100 random coordinates in a 31×31×31 cube
2. Groups co-located points into systems with multiple stars
3. For each star, generates orbits (weighted random: 29% terrestrial, 5% asteroid belt, 7% gas giant, 59% empty; max 3 gas giants, max 2 asteroid belts; sorted by type)
4. For each occupied orbit, generates a planet with habitability (orbit-dependent curves) and 0–34 resource deposits (type/quantity/yield from weighted tables)
5. Normalizes the tree into a flat `domain.Cluster` with shuffled planet and deposit IDs

The generation tables are based on the 1978 rules.

**Acceptance criteria:**
- [x] `go build ./internal/domain/clustergen/...` succeeds
- [x] `GenerateCluster` returns a valid `domain.Cluster` with systems, stars, planets, and deposits
- [x] Same seeds produce identical output (deterministic)
- [x] All internal types are unexported
- [x] Only imports `domain` and `prng` (plus stdlib)

**Tests to add/update:**
- None yet — determinism validated via the distribution tester in Task 4

---

### Task 3: Cluster stats aggregation

**Subsystem:** `domain/clustergen`
**Files:** `backend/internal/domain/clustergen/stats.go`
**Depends on:** Tasks 1, 2

**What to do:**
Implement `ClusterStats` for aggregating generation statistics over `domain.Cluster`. Used by the distribution tester to validate generation tables across many runs.

Types: `DepositStats` (Count, TotalQty, TotalPct), `ClusterStats` (NumSystems, NumStars, TotalPlanets, PlanetsByKind, HabitableByKind, TotalHabByKind, Overall deposits, ByPlanetKind deposits).

`NewClusterStats()` initializes all maps. `Collect(domain.Cluster)` walks the flat cluster model, building a deposit-by-ID lookup and aggregating counts.

**Acceptance criteria:**
- [x] `go build ./internal/domain/clustergen/...` succeeds
- [x] `Collect` accepts `domain.Cluster` (not internal tree types)
- [x] All fields on `DepositStats` are exported (needed by delivery for report formatting)

**Tests to add/update:**
- None — validated indirectly via the distribution tester

---

### Task 4: App-layer ports and ClusterService

**Subsystem:** `app`
**Files:** `backend/internal/app/cluster_ports.go`, `backend/internal/app/cluster_service.go`
**Depends on:** Tasks 2, 3

**What to do:**
Define port interfaces for cluster persistence and implement the `ClusterService` with three use cases.

Ports:
- `ClusterReader` — `ReadCluster(path string) (domain.Cluster, error)`
- `ClusterWriter` — `WriteCluster(path string, cluster domain.Cluster) error`
- `GameWriter` — `WriteGame(path string, game *domain.Game, overwrite bool) error`

Use cases on `ClusterService`:
- `CreateCluster(seed1, seed2 uint64, outputPath string) (domain.Cluster, error)` — generates one cluster, writes it, returns it for reporting
- `TestCluster(seed1, seed2 uint64, iterations int) (*clustergen.ClusterStats, error)` — runs N iterations, returns aggregated stats (no file I/O)
- `CreateGame(clusterPath, savePath string, overwrite bool) error` — reads a cluster file, wraps it in a `domain.Game`, writes the game file

**Acceptance criteria:**
- [x] `go build ./internal/app/...` succeeds
- [x] No imports of Echo, Cobra, SQLite, or filesystem packages
- [x] `ClusterService` depends only on port interfaces, not concrete types
- [x] PRNG construction is centralized in the service (not duplicated in callers)

**Tests to add/update:**
- None — use cases validated via CLI integration

---

### Task 5: Infra filestore adapter for clusters and games

**Subsystem:** `infra/filestore`
**Files:** `backend/internal/infra/filestore/cluster.go`
**Depends on:** Task 4

**What to do:**
Add methods to the existing `filestore.Store` that implement `app.ClusterReader`, `app.ClusterWriter`, and `app.GameWriter`.

- `ReadCluster(path)` — reads and unmarshals `domain.Cluster` from JSON
- `WriteCluster(path, cluster)` — marshals and writes `domain.Cluster` as indented JSON, creating parent directories as needed
- `WriteGame(path, game, overwrite)` — marshals and writes `*domain.Game` as indented JSON; returns error if file exists and `overwrite` is false

**Acceptance criteria:**
- [x] `go build ./internal/infra/filestore/...` succeeds
- [x] `Store` satisfies all three port interfaces
- [x] `WriteGame` with `overwrite=false` returns error on existing file
- [x] No imports of `delivery` or `runtime`

**Tests to add/update:**
- None — existing filestore tests continue to pass; cluster adapter validated via CLI integration

---

### Task 6: CLI delivery commands and report formatting

**Subsystem:** `delivery/cli`
**Files:** `backend/internal/delivery/cli/cluster.go`, `backend/internal/delivery/cli/report.go`
**Depends on:** Tasks 3, 4

**What to do:**
Create thin cobra command builders that receive `*app.ClusterService` and delegate to it. Also implement report formatting (moved from the original generators package).

Commands:
- `CmdCreateCluster(svc)` — flags: `--path`, `--seed1`, `--seed2`. Calls `svc.CreateCluster`, prints report via `WriteClusterReport`.
- `CmdTestCluster(svc)` — flags: `--iterations`, `--seed1`, `--seed2`. Calls `svc.TestCluster`, prints report via `WriteStatsReport`.
- `CmdCreateGame(svc)` — flags: `--cluster`, `--save`, `--overwrite`. Calls `svc.CreateGame`.

Report functions:
- `WriteClusterReport(w, cluster)` — collects stats from one cluster and prints a report
- `WriteStatsReport(w, stats, divisor)` — formats the full stats table with planet counts, habitability averages, deposit quantities, and yield percentages per resource and planet type
- `commaFmtInt64` — number formatting helper

**Acceptance criteria:**
- [x] `go build ./internal/delivery/cli/...` succeeds
- [x] Handlers are thin — no generation logic, no file I/O, no PRNG construction
- [x] No imports of `infra` or `runtime`
- [x] Report output matches the original format from `generators/cluster.go`

**Tests to add/update:**
- None — delivery validated via CLI integration

---

### Task 7: Runtime CLI wiring

**Subsystem:** `runtime/cli`
**Files:** `backend/internal/runtime/cli/cli.go`
**Depends on:** Tasks 5, 6

**What to do:**
Create `AddCommands(root *cobra.Command)` that wires concrete infra into app services and attaches delivery commands to the root cobra command.

1. Instantiate `filestore.NewStore("")`
2. Create `app.ClusterService` with store as Reader, Writer, and GameWriter
3. Build `create` command group with `CmdCreateCluster` and `CmdCreateGame`
4. Build `test` command group with `CmdTestCluster`
5. Add both groups to root

**Acceptance criteria:**
- [x] `go build ./internal/runtime/cli/...` succeeds
- [x] `runtime/cli` is the only package that imports both `infra` and `delivery`
- [x] `cmd/cli/main.go` calls `runtimecli.AddCommands(cmdRoot)` — no inline command construction

**Tests to add/update:**
- None — wiring validated via CLI integration

---

### Task 8: Eliminate orphan packages

**Subsystem:** cleanup
**Files:** Deleted: `backend/internal/generators/`, `backend/internal/adapters/`, `backend/internal/fsck/`. Modified: `backend/internal/dotfiles/dotfiles.go`, `backend/cmd/api/main.go`.
**Depends on:** Tasks 1–7

**What to do:**
Remove the orphan packages that existed outside the SOUSA layer structure:
- `internal/generators/` — generation logic moved to `domain/clustergen`
- `internal/adapters/` — normalization moved to `clustergen`, file I/O moved to `infra/filestore`
- `internal/fsck/` — `IsFile`/`IsDir` inlined at call sites

Update remaining references:
- `internal/dotfiles/dotfiles.go` — replace `fsck.IsFile` with a local `isFile` helper using `os.Stat`
- `cmd/api/main.go` — replace `fsck.IsDir` with inline `os.Stat` check

**Acceptance criteria:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` is clean
- [x] No remaining imports of `generators`, `adapters`, or `fsck`
- [x] No packages exist outside the SOUSA layer structure

**Tests to add/update:**
- Existing tests in `infra/filestore`, `infra/auth`, `delivery/http`, `runtime/server` continue to pass

---

## Post-Sprint Review Findings

A SOUSA compliance and code-smell review was performed before any code was committed. The following issues were identified and fixed as part of this sprint.

### SOUSA Violations (fixed)

| # | Finding | Fix |
|---|---------|-----|
| V1 | `generators/` package outside SOUSA layer structure | Moved to `domain/clustergen/` |
| V2 | `adapters/` package outside SOUSA layer structure — mixed file I/O with domain conversion | Split: normalization into `clustergen`, file I/O into `infra/filestore`, conversion into `app` |
| V3 | `fsck/` utility package outside SOUSA layer structure | Eliminated; inlined `os.Stat` at call sites |
| V4 | `adapters` imported `generators` and did file I/O — circular-risk dependency mixing concerns | Eliminated by splitting across proper layers |
| V5 | `cmd/cli/main.go` embedded generation logic, JSON serialization, and file writes in cobra handlers | Decomposed into `domain/clustergen` → `app` → `infra/filestore` → `delivery/cli` → `runtime/cli` |
| V6 | Generator tree types were exported, leaking internal representation to outer layers | Made all tree types unexported; only `GenerateCluster` is public |

### Code Smells (fixed)

| # | Finding | Fix |
|---|---------|-----|
| S1 | `commaFmtInt` defined but never called | Deleted |
| S2 | `ClusterStats.Add` defined but never called | Deleted |
| S3 | PRNG construction (`prng.New(rand.NewPCG(...))`) duplicated in two commands | Centralized in `ClusterService` methods |
| S4 | Mixed `log.Fatal` and `slog` error handling in `cmd/cli/main.go` | Removed `log` import; all error handling uses `slog` or cobra's `RunE` |
| S5 | Raw `json.MarshalIndent` + `os.WriteFile` in cobra handler duplicated pattern from `adapters.GameToJson` | Consolidated in `infra/filestore` methods |
| S6 | Generator defined `System`/`Star`/`Planet`/`Deposit` types that duplicated `domain` types | Tree types made unexported; generator returns `domain.Cluster` directly |

---

## Task Summary

| Task | Title                                    | Status | Agent/Thread | Notes |
|------|------------------------------------------|--------|--------------|-------|
| 1    | Domain types for clusters                | DONE   |              | Retroactive |
| 2    | Cluster generation logic                 | DONE   |              | Retroactive |
| 3    | Cluster stats aggregation                | DONE   |              | Retroactive |
| 4    | App-layer ports and ClusterService       | DONE   |              | Retroactive |
| 5    | Infra filestore adapter                  | DONE   |              | Retroactive |
| 6    | CLI delivery commands and reporting      | DONE   |              | Retroactive |
| 7    | Runtime CLI wiring                       | DONE   |              | Retroactive |
| 8    | Eliminate orphan packages                | DONE   |              | Retroactive |
