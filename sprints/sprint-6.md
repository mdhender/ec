# Sprint 6: Replace Cobra + godotenv with ff/v4

**Pass:** Pass 1
**Goal:** Replace spf13/cobra, joho/godotenv, and hand-rolled resolver helpers with peterbourgon/ff/v4, eliminating duplicated flag/env plumbing and fixing broken flag bindings.
**Predecessor:** Sprint 5

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

This sprint replaces the CLI framework and configuration plumbing. The project
currently uses spf13/cobra for command trees, joho/godotenv for .env file
loading, and hand-rolled `resolveString`/`resolveDuration` helpers (duplicated
in 4 files) for flag → env → fallback resolution.

The replacement is **peterbourgon/ff/v4**, which provides:
- `ff.Command` — declarative command trees (replaces cobra)
- `ff.FlagSet` — flag definitions with short/long names
- `ff.Parse` — unified parsing with built-in flag → env → config file → default priority
- `ff.WithEnvVarPrefix("EC")` — auto-maps `--data-path` to `EC_DATA_PATH`
- `ffenv.Parse` — .env file parser (replaces godotenv)
- `ffhelp` — help text formatting

**Key design decisions:**
- Env prefix is `EC` everywhere (e.g., `--data-path` maps to `EC_DATA_PATH`)
- Rename `--path` to `--data-path` on CLI commands for consistency with `EC_DATA_PATH`
- Validate required values after parse (`if val == "" { return err }`), not via framework mechanisms — this lets env vars and .env files satisfy "required" fields
- Logs go to stderr, command output goes to stdout
- Remove `--info` dead flag; keep `--debug` and `--quiet` as convenience aliases resolved before ff.Parse

**Known bugs being fixed:**
1. `CmdShowMagicLink` defines `--data-path` but resolves `"path"` — the flag is silently ignored
2. `--info` flag is defined but never read in `PersistentPreRunE`
3. `resolveString`/`resolveDuration` are copy-pasted in 4 files

After completing a task, update sprints/sprint-6.md: check off acceptance criteria (change [ ] to [x]) and change the task status from TODO to DONE in the Task Summary table at the bottom of the file.

**Key files (before migration):**
- `backend/go.mod` — add ff/v4, remove cobra + godotenv
- `backend/cmd/api/main.go` — API server entry point (cobra root, resolve helpers, serve command)
- `backend/cmd/cli/main.go` — CLI entry point (cobra root, resolve helpers, dotfiles.Load)
- `backend/internal/dotfiles/dotfiles.go` — .env loading logic (to be deleted)
- `backend/internal/delivery/cli/resolver.go` — resolve helpers (to be deleted)
- `backend/internal/delivery/cli/cluster.go` — cobra commands for cluster operations
- `backend/internal/delivery/cli/game_config.go` — cobra commands for game config
- `backend/internal/runtime/cli/cli.go` — CLI wiring (cobra command tree + resolve helpers)
- `backend/internal/runtime/server/options.go` — server functional options

**Key files (after migration):**
- `backend/cmd/api/main.go` — ff.Command root, ff.Parse with WithEnvVarPrefix
- `backend/cmd/cli/main.go` — ff.Command root, ff.Parse with WithEnvVarPrefix + ffenv
- `backend/internal/delivery/cli/cluster.go` — ff.Command/ff.FlagSet commands
- `backend/internal/delivery/cli/game_config.go` — ff.Command/ff.FlagSet commands
- `backend/internal/runtime/cli/cli.go` — wiring (ff.Command tree, no resolve helpers)

**Build/test commands:**
```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/api/
cd backend && go build ./cmd/cli/
```

**Constraints reminder:**
- `domain` must not import `app`, `infra`, `delivery`, or `runtime`
- `app` must not import Echo, SQLite, CLI frameworks, or filesystem adapters
- `delivery` must not import `infra` — they are peers
- `runtime` is the only layer that wires concrete implementations

---

## Tasks

### Task 1: Add ff/v4 dependency, remove cobra and godotenv

**Subsystem:** build / go.mod
**Files:** `backend/go.mod`
**Depends on:** None

**What to do:**
In the `backend/` directory:
1. Run `go get github.com/peterbourgon/ff/v4@latest`
2. Run `go get github.com/spf13/cobra@none` (removes cobra)
3. Run `go get github.com/spf13/pflag@none` (removes pflag, cobra's transitive dep)
4. Run `go get github.com/joho/godotenv@none` (removes godotenv)
5. Run `go get github.com/inconshreveable/mousetrap@none` (removes mousetrap, cobra's transitive dep)
6. Run `go mod tidy`

**Important:** This task only changes `go.mod` and `go.sum`. The code will not
compile after this task until subsequent tasks update the Go source files.
That is expected — this task intentionally does **not** touch `.go` files.

**Acceptance criteria:**
- [x] `go.mod` lists `github.com/peterbourgon/ff/v4`
- [x] `go.mod` does not list `github.com/spf13/cobra`, `github.com/spf13/pflag`, `github.com/joho/godotenv`, or `github.com/inconshreveable/mousetrap`

**Tests to add/update:**
- None — dependency management only

---

### Task 2: Delete dotfiles package and resolver helpers

**Subsystem:** cleanup
**Files:**
- `backend/internal/dotfiles/dotfiles.go` (delete)
- `backend/internal/delivery/cli/resolver.go` (delete)
**Depends on:** None (parallel with Task 1; code won't compile until later tasks)

**What to do:**
1. Delete `backend/internal/dotfiles/dotfiles.go` — the entire `dotfiles`
   package is replaced by `ff.WithConfigFile` + `ffenv.Parse`.
2. Delete `backend/internal/delivery/cli/resolver.go` — the `resolveString`
   and `resolveDuration` helpers are replaced by `ff.Parse` with
   `WithEnvVarPrefix`.

These files have no tests of their own. All callers will be updated in
Tasks 3–5.

**Acceptance criteria:**
- [x] `backend/internal/dotfiles/` directory no longer exists
- [x] `backend/internal/delivery/cli/resolver.go` no longer exists

**Tests to add/update:**
- None — deleted code had no tests

---

### Task 3: Migrate cmd/api to ff/v4

**Subsystem:** `cmd/api`
**Files:** `backend/cmd/api/main.go`
**Depends on:** Tasks 1, 2

**What to do:**
Rewrite `cmd/api/main.go` to use `ff.Command`, `ff.FlagSet`, and `ff.Parse`
instead of cobra. Remove the `resolveString`, `resolveDuration` functions and
the `dotfiles.Load` call.

The resulting structure should be:

1. **Root command** (`api`):
   - FlagSet with parent-level flags: `--log-level` (string, default `"info"`),
     `--log-source` (bool), `--debug` (bool), `--quiet` (bool).
   - Remove the dead `--info` flag entirely.
   - No Exec — root just dispatches to subcommands.

2. **`serve` subcommand**:
   - FlagSet (parent: root flagset) with: `--host` (default `"localhost"`),
     `--port` (default `"8080"`), `--data-path`, `--jwt-secret`,
     `--shutdown-key`, `--timeout` (duration, default 0).
   - Exec function: read flag values, validate required fields (`data-path`,
     `jwt-secret`), configure logger from log flags, create and start server.
   - Post-parse validation: if `dataPath == ""` return error; if `jwtSecret == ""`
     return error; `os.Stat(dataPath)` must be a directory.

3. **`show` command group** with **`version`** subcommand:
   - `--build-info` flag (bool).
   - Prints `ec.Version().String()` or `ec.Version().Core()`.

4. **Parse options** (applied at root level):
   - `ff.WithEnvVarPrefix("EC")` — auto-maps `--data-path` → `EC_DATA_PATH`, etc.
   - `ff.WithConfigFile(".env")`, `ff.WithConfigFileParser(ffenv.Parse)`,
     `ff.WithConfigAllowMissingFile()` — loads `.env` if present.
   - `ff.WithConfigIgnoreFlagNames()` — .env keys match env var names only.

5. **Logging setup**: create the slog handler writing to `os.Stderr` (not
   stdout). Resolve `--debug`/`--quiet` as overrides of `--log-level`:
   - If `--debug` is true, force level to debug.
   - If `--quiet` is true, force level to error.
   - If both, return error.
   - Otherwise parse `--log-level` string.

6. **Error handling**: on error, print help via `ffhelp.Command`, print the
   error to stderr, and `os.Exit(1)`.

**Reference ff.Command pattern:**
```go
rootFlags := ff.NewFlagSet("api")
logLevel := rootFlags.StringLong("log-level", "info", "log level (debug|info|warn|error)")
// ...

serveFlags := ff.NewFlagSet("serve").SetParent(rootFlags)
host := serveFlags.StringLong("host", "localhost", "listen host")
// ...

serveCmd := &ff.Command{
    Name:  "serve",
    Usage: "api serve [FLAGS]",
    Flags: serveFlags,
    Exec:  func(ctx context.Context, args []string) error { ... },
}

rootCmd := &ff.Command{
    Name:        "api",
    Usage:       "api [FLAGS] SUBCOMMAND ...",
    Flags:       rootFlags,
    Subcommands: []*ff.Command{serveCmd, showCmd},
}

err := rootCmd.ParseAndRun(ctx, os.Args[1:],
    ff.WithEnvVarPrefix("EC"),
    ff.WithConfigFile(".env"),
    ff.WithConfigFileParser(ffenv.Parse),
    ff.WithConfigAllowMissingFile(),
    ff.WithConfigIgnoreFlagNames(),
)
```

**Acceptance criteria:**
- [x] `cd backend && go build ./cmd/api/` succeeds
- [x] No imports of `cobra`, `pflag`, `godotenv`, or `dotfiles`
- [x] `--data-path` flag maps to `EC_DATA_PATH` env var automatically
- [x] `--info` flag is removed
- [x] Logger writes to stderr, not stdout
- [x] `api serve --help` shows all flags with defaults
- [x] `api serve` without required flags prints an error

**Tests to add/update:**
- None — validated via build and manual smoke test

---

### Task 4: Migrate delivery/cli commands to ff/v4

**Subsystem:** `delivery/cli`
**Files:**
- `backend/internal/delivery/cli/cluster.go`
- `backend/internal/delivery/cli/game_config.go`
**Depends on:** Task 1

**What to do:**
Convert all command-builder functions from returning `*cobra.Command` to
returning `*ff.Command`. Each function should create an `ff.FlagSet`, define
its flags, and return an `ff.Command` with an `Exec` function.

**Cluster commands** (`cluster.go`):

1. `CmdCreateCluster(svc *app.ClusterService) *ff.Command`:
   - FlagSet `"cluster"` (no parent — parent set by caller in runtime wiring).
   - Flags: `--path` (string, default `"testdata/cluster.json"`), `--seed1`
     (uint64? — ff doesn't have Uint64; use `StringLong` and parse with
     `strconv.ParseUint`, or define a custom `ffval` value), `--seed2`,
     `--overwrite` (bool).
   - **Note on uint64 flags:** `ff.FlagSet` does not have a native `Uint64`
     method. Options: (a) use `StringLong` and `strconv.ParseUint` in Exec,
     or (b) use `ffval.NewValueDefault[uint64](strconv.ParseUint wrapper)`.
     Choose option (a) for simplicity.
   - Exec: parse seed strings to uint64, call `svc.CreateCluster(...)`,
     write report to stdout.

2. `CmdTestCluster(svc *app.ClusterService) *ff.Command`:
   - Flags: `--iterations` (int, default 100), `--seed1`, `--seed2` (same
     uint64 approach as above).
   - Exec: call `svc.TestCluster(...)`, write stats report to stdout.

3. `CmdCreateGameState(svc *app.ClusterService) *ff.Command`:
   - Flags: `--cluster` (string, required), `--save` (string, required),
     `--overwrite` (bool).
   - Exec: validate required flags, call `svc.CreateGame(...)`.

**Game config commands** (`game_config.go`):

4. `CmdCreateGame(svc *app.GameConfigService) *ff.Command`:
   - Flags: `--data-path` (string, required). **Renamed from `--path`.**
   - Exec: validate `data-path` not empty, call `svc.CreateGame(dataPath)`.

5. `CmdAddEmpire(svc *app.GameConfigService) *ff.Command`:
   - Flags: `--data-path` (string, required), `--empire` (int, default 0).
   - Exec: validate, call `svc.AddEmpire(dataPath, empireNo)`.

6. `CmdRemoveEmpire(svc *app.GameConfigService) *ff.Command`:
   - Flags: `--data-path` (string, required), `--empire` (int, required).
   - Exec: validate both not empty/zero, call `svc.RemoveEmpire(...)`.

7. `CmdShowMagicLink(svc *app.GameConfigService) *ff.Command`:
   - Flags: `--data-path` (string, required), `--base-url` (string, default
     from empty — will be populated from `EC_BASE_URL` via env prefix),
     `--empire` (int, required).
   - Exec: validate, call `svc.ShowMagicLink(dataPath, empireNo)`, format
     and print URL.
   - **This fixes the bug** where the old code defined `--data-path` but
     resolved `"path"`.

Remove all imports of `cobra` and `pflag` from these files. Remove any
`resolveString` calls — ff handles flag → env → default automatically.

**Acceptance criteria:**
- [x] `cd backend && go build ./internal/delivery/cli/...` succeeds
- [x] No imports of `cobra`, `pflag`, or `resolver.go` helpers
- [x] All commands use `--data-path` (not `--path`) for the data directory flag
- [x] `CmdShowMagicLink` correctly reads `--data-path` (bug fix verified)
- [x] Commands are thin — no business logic, no file I/O beyond what svc provides
- [x] No imports of `infra` or `runtime`

**Tests to add/update:**
- None — delivery validated via CLI integration

---

### Task 5: Migrate runtime/cli wiring to ff/v4

**Subsystem:** `runtime/cli`
**Files:** `backend/internal/runtime/cli/cli.go`
**Depends on:** Tasks 1, 4

**What to do:**
Update `AddCommands` to build the ff.Command tree instead of the cobra tree.
Change the function signature to return the list of subcommands rather than
mutating a cobra root:

```go
func BuildCommands() []*ff.Command
```

Or, if simpler, keep a similar signature that takes the parent FlagSet:

```go
func BuildCommands(parentFlags *ff.FlagSet) []*ff.Command
```

The function should:
1. Create `filestore.NewStore("")` and service instances (same as today).
2. Build command groups:
   - `create` group containing: `cluster`, `game-state`, `game`, `empire`
   - `remove` group containing: `empire`
   - `show` group containing: `magic-link`
   - `test` group containing: `cluster`
3. Return the command list for the caller (`cmd/cli/main.go`) to attach.

Remove the `version` command from this file — it moves to `cmd/cli/main.go`
(or `cmd/api/main.go`) where the root command is defined.

Delete the `resolveString` and `resolveDuration` functions from this file.

**Acceptance criteria:**
- [x] `cd backend && go build ./internal/runtime/cli/...` succeeds
- [x] No imports of `cobra` or `pflag`
- [x] No `resolveString` or `resolveDuration` functions remain
- [x] `version` command is no longer defined here (moved to cmd entry points)

**Tests to add/update:**
- None — wiring validated via CLI integration

---

### Task 6: Migrate cmd/cli to ff/v4

**Subsystem:** `cmd/cli`
**Files:** `backend/cmd/cli/main.go`
**Depends on:** Tasks 1, 2, 5

**What to do:**
Rewrite `cmd/cli/main.go` to use `ff.Command` and `ff.Parse`. This is similar
to Task 3 (cmd/api) but uses `runtime/cli.BuildCommands()` for the command
tree.

1. **Root command** (`cli`):
   - FlagSet with: `--log-level`, `--log-source`, `--debug`, `--quiet`.
   - Remove the dead `--info` flag.

2. **Subcommands**: call `runtimecli.BuildCommands(rootFlags)` (or equivalent)
   and attach the returned commands as subcommands of root.

3. **`version` subcommand**: define inline (same as Task 3's pattern).

4. **Parse options** (same as Task 3):
   - `ff.WithEnvVarPrefix("EC")`
   - `ff.WithConfigFile(".env")`, `ff.WithConfigFileParser(ffenv.Parse)`,
     `ff.WithConfigAllowMissingFile()`, `ff.WithConfigIgnoreFlagNames()`

5. **Logging**: same stderr-based setup as Task 3.

6. **Remove**: `dotfiles.Load` call, `resolveString`, `resolveDuration`
   functions, all cobra/pflag imports.

**Acceptance criteria:**
- [x] `cd backend && go build ./cmd/cli/` succeeds
- [x] No imports of `cobra`, `pflag`, `godotenv`, or `dotfiles`
- [x] `--info` flag is removed
- [x] Logger writes to stderr, not stdout
- [x] `cli --help` shows command tree
- [x] `cli create cluster --help` shows cluster flags
- [x] All existing tests pass: `cd backend && go test ./...`

**Tests to add/update:**
- None — validated via build and smoke test

---

### Task 7: Verify build, clean up dead code, run full test suite

**Subsystem:** all
**Files:** any remaining references
**Depends on:** Tasks 3, 4, 5, 6

**What to do:**
1. Run `cd backend && go build ./...` — fix any remaining compilation errors.
2. Run `cd backend && go test ./...` — fix any test failures.
3. Search for any remaining references to removed packages:
   - `grep -r "godotenv\|spf13/cobra\|spf13/pflag\|dotfiles" backend/`
   - `grep -r "resolveString\|resolveDuration" backend/`
   - Fix or remove any found.
4. Delete `backend/internal/dotfiles/` directory if not already removed.
5. Run `go mod tidy` to clean up any unused transitive dependencies.
6. Verify `go vet ./...` passes.

**Acceptance criteria:**
- [x] `cd backend && go build ./...` succeeds with zero errors
- [x] `cd backend && go test ./...` passes all tests
- [x] `cd backend && go vet ./...` passes
- [x] No source files reference `cobra`, `pflag`, `godotenv`, or `dotfiles`
- [x] No `resolveString` or `resolveDuration` functions exist anywhere
- [x] `go.mod` has no unused dependencies

**Tests to add/update:**
- `TestEnvPrecedence` in `backend/cmd/cli/main_test.go` (new file) — a minimal
  table-driven test that verifies ff parse priority: flag value beats env var
  beats .env file beats default. Use `ff.Parse` with a temp .env file and
  `t.Setenv` to exercise the chain. At minimum test one string flag and one
  duration flag.

---

## Task Summary

| Task | Title                                          | Status | Agent/Thread | Notes |
|------|------------------------------------------------|--------|--------------|-------|
| 1    | Add ff/v4 dependency, remove cobra + godotenv  | DONE   |              |       |
| 2    | Delete dotfiles package and resolver helpers   | DONE   |              |       |
| 3    | Migrate cmd/api to ff/v4                       | DONE   |              |       |
| 4    | Migrate delivery/cli commands to ff/v4         | DONE   |              |       |
| 5    | Migrate runtime/cli wiring to ff/v4            | DONE   |              |       |
| 6    | Migrate cmd/cli to ff/v4                       | DONE   |              |       |
| 7    | Verify build, clean up dead code, full tests   | DONE   |              |       |
