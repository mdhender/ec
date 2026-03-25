# Sprint 5: Game, Empire, and Auth CLI Commands

**Pass:** Pass 1
**Goal:** Add CLI commands to create a game, add/remove empires, and show magic link URLs — managing `game.json` and `auth.json` files on disk.
**Predecessor:** Sprint 4

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

This sprint adds four CLI commands for managing game setup files. These files
are simple JSON configs — they are **not** the full domain `Game`/`Empire`
types (which include clusters, colonies, ships, etc.). The setup files track
which empires are registered and how they authenticate.

The sample files in `backend/data/alpha/` show the on-disk format:

**`game.json`** — list of empires with active status:
```json
{
  "empires": [
    {"empire": 42, "active": true},
    {"empire": 1812, "active": true}
  ]
}
```

**`auth.json`** — map of UUID magic links to empire numbers:
```json
{
  "magic-links": {
    "37e81785-84ee-4fee-850b-160e373a4539": {"empire": 42},
    "81ce2bb6-42fe-49b2-80c5-0558787c8471": {"empire": 1812}
  }
}
```

The project follows SOUSA layering (see `docs/SOUSA.md`). Dependencies flow
inward only: `domain ← app ← infra / delivery ← runtime`.

After completing a task, update sprints/sprint-5.md: check off acceptance
criteria (change [ ] to [x]) and change the task status from TODO to DONE in
the Task Summary table at the bottom of the file.

**Key files:**
- `backend/internal/domain/game_config.go` — new: `GameConfig`, `EmpireEntry`, `AuthConfig` types
- `backend/internal/app/game_config_ports.go` — new: `GameConfigStore` port interface
- `backend/internal/app/game_config_service.go` — new: `GameConfigService` use cases
- `backend/internal/infra/filestore/game_config.go` — new: JSON read/write for game & auth config files
- `backend/internal/delivery/cli/game_config.go` — new: thin cobra commands
- `backend/internal/runtime/cli/cli.go` — wire new service and commands
- `backend/internal/infra/auth/magiclinks.go` — existing auth.json loader (read-only; reference for JSON format)
- `backend/data/alpha/game.json`, `backend/data/alpha/auth.json` — sample files

**Key types/functions (to be created):**
- `domain.GameConfig` — `Empires []EmpireEntry`
- `domain.EmpireEntry` — `Empire int`, `Active bool`
- `domain.AuthConfig` — `MagicLinks map[string]AuthLink`
- `domain.AuthLink` — `Empire int`
- `app.GameConfigStore` — port interface for reading/writing game.json and auth.json
- `app.GameConfigService` — orchestrates the four use cases

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
- stdlib first — minimal third-party dependencies

---

## Tasks

### Task 1: Domain types for game and auth config

**Subsystem:** `domain`
**Files:** `backend/internal/domain/game_config.go` (new file)
**Depends on:** None

**What to do:**
Create lightweight domain types for the game setup files. These are separate
from the existing `domain.Game` and `domain.Empire` which represent full game
state. These types represent the on-disk config format:

```go
// GameConfig is the on-disk structure for game.json.
type GameConfig struct {
    Empires []EmpireEntry `json:"empires"`
}

// EmpireEntry is one empire's registration in game.json.
type EmpireEntry struct {
    Empire int  `json:"empire"`
    Active bool `json:"active"`
}

// AuthConfig is the on-disk structure for auth.json.
type AuthConfig struct {
    MagicLinks map[string]AuthLink `json:"magic-links"`
}

// AuthLink maps a magic link UUID to an empire number.
type AuthLink struct {
    Empire int `json:"empire"`
}
```

These types must have JSON tags matching the sample files in `backend/data/alpha/`.

**Acceptance criteria:**
- [ ] `cd backend && go build ./internal/domain/...` succeeds
- [ ] Types have JSON tags: `GameConfig.Empires` → `"empires"`, `EmpireEntry.Empire` → `"empire"`, `EmpireEntry.Active` → `"active"`, `AuthConfig.MagicLinks` → `"magic-links"`, `AuthLink.Empire` → `"empire"`
- [ ] No imports of `app`, `infra`, `delivery`, or `runtime`

**Tests to add/update:**
- None — pure type definitions

---

### Task 2: App-layer port and GameConfigService

**Subsystem:** `app`
**Files:** `backend/internal/app/game_config_ports.go` (new), `backend/internal/app/game_config_service.go` (new)
**Depends on:** Task 1

**What to do:**

Define a port interface for reading/writing game and auth config files:

```go
// GameConfigStore reads and writes game.json and auth.json files.
type GameConfigStore interface {
    ReadGameConfig(path string) (domain.GameConfig, error)
    WriteGameConfig(path string, cfg domain.GameConfig) error
    ReadAuthConfig(path string) (domain.AuthConfig, error)
    WriteAuthConfig(path string, cfg domain.AuthConfig) error
}
```

The `path` parameter is the **directory** containing `game.json` and `auth.json`.
The store implementation appends the filename internally.

Create `GameConfigService` with four methods:

1. **`CreateGame(dirPath string) error`**
   - Verify `dirPath` exists and is a directory.
   - Fail if `game.json` or `auth.json` already exists in that directory.
   - Write an empty `GameConfig{Empires: []EmpireEntry{}}` to `game.json`.
   - Write an empty `AuthConfig{MagicLinks: map[string]AuthLink{}}` to `auth.json`.

2. **`AddEmpire(dirPath string, empireNo int) error`**
   - Read `game.json` from `dirPath`.
   - If `empireNo` is 0, auto-assign: find the largest empire number in the
     list and add 1 (or start at 1 if the list is empty).
   - Fail if an empire with that number already exists (active or not).
   - Append `EmpireEntry{Empire: empireNo, Active: true}` to the list.
   - Write updated `game.json`.
   - Read `auth.json`, generate a new UUID (use `crypto/rand` via
     `uuid.New()` or format manually), add a new magic link entry mapping
     the UUID to the empire number. Write updated `auth.json`.
   - Return the assigned empire number and generated magic link UUID to the
     caller (adjust return signature: `(empireNo int, magicLink string, err error)`).

3. **`RemoveEmpire(dirPath string, empireNo int) error`**
   - Read `game.json` from `dirPath`.
   - Fail if no empire with that number exists.
   - Set `Active = false` on the matching entry. Write updated `game.json`.
   - Read `auth.json`. Remove any magic link entries for that empire number.
     If there are no matching links, that's fine — no error. Write updated
     `auth.json`.

4. **`ShowMagicLink(dirPath string, empireNo int) (string, error)`**
   - Read `auth.json` from `dirPath`.
   - Find the magic link UUID for the given empire number.
   - Return the full URL: `https://app.epimethean.dev/?magic={uuid}`.
   - Fail if no magic link exists for that empire.

For UUID generation: use `crypto/rand` to generate a v4 UUID. Format it as
`xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`. Do **not** add a third-party UUID
library — implement a small helper in the service file or in a new
`domain` helper.

**Acceptance criteria:**
- [ ] `cd backend && go build ./internal/app/...` succeeds
- [ ] No imports of Echo, Cobra, SQLite, or filesystem packages
- [ ] `GameConfigService` depends only on the `GameConfigStore` interface
- [ ] `CreateGame` fails if either `game.json` or `auth.json` already exists
- [ ] `AddEmpire` auto-assigns empire number when input is 0
- [ ] `AddEmpire` fails if empire already exists
- [ ] `RemoveEmpire` sets active=false, does not delete the empire entry
- [ ] `RemoveEmpire` silently succeeds when no magic link exists for the empire
- [ ] `ShowMagicLink` returns full URL with `https://app.epimethean.dev/?magic=` prefix

**Tests to add/update:**
- `TestCreateGame` in `backend/internal/app/game_config_service_test.go` — use a mock `GameConfigStore` to verify:
  - calls `WriteGameConfig` with empty empires list
  - calls `WriteAuthConfig` with empty magic links map
  - returns error if store reports file exists
- `TestAddEmpire` in same file — verify auto-numbering logic (empty list → 1, list with [3,7] → 8), duplicate detection, magic link generation
- `TestRemoveEmpire` in same file — verify active flag set to false, magic link removed, no error when link missing

---

### Task 3: Infra filestore adapter for game and auth config

**Subsystem:** `infra/filestore`
**Files:** `backend/internal/infra/filestore/game_config.go` (new)
**Depends on:** Task 2

**What to do:**
Add methods to the existing `filestore.Store` that implement `app.GameConfigStore`.
The `path` parameter for all methods is a **directory path**. The store appends
`game.json` or `auth.json` internally.

```go
func (s *Store) ReadGameConfig(dirPath string) (domain.GameConfig, error)
func (s *Store) WriteGameConfig(dirPath string, cfg domain.GameConfig) error
func (s *Store) ReadAuthConfig(dirPath string) (domain.AuthConfig, error)
func (s *Store) WriteAuthConfig(dirPath string, cfg domain.AuthConfig) error
```

- `ReadGameConfig`: reads `filepath.Join(dirPath, "game.json")`, unmarshals into `domain.GameConfig`. Returns error if file doesn't exist or JSON is invalid.
- `WriteGameConfig`: marshals `domain.GameConfig` as indented JSON, writes to `filepath.Join(dirPath, "game.json")`. Overwrites if file exists (the service layer handles existence checks).
- `ReadAuthConfig` / `WriteAuthConfig`: same pattern for `auth.json` and `domain.AuthConfig`.

Write operations should use `json.MarshalIndent(cfg, "", "  ")` with a trailing
newline for readability.

**Acceptance criteria:**
- [ ] `cd backend && go build ./internal/infra/filestore/...` succeeds
- [ ] `Store` satisfies `app.GameConfigStore` interface
- [ ] Read methods return meaningful errors on missing file or bad JSON
- [ ] Written JSON matches the format in `backend/data/alpha/` sample files
- [ ] No imports of `delivery` or `runtime`

**Tests to add/update:**
- `TestGameConfigRoundTrip` in `backend/internal/infra/filestore/game_config_test.go` — write a `GameConfig` with two empires to a temp dir, read it back, verify equality
- `TestAuthConfigRoundTrip` in same file — write an `AuthConfig` with two magic links to a temp dir, read it back, verify equality
- `TestReadGameConfigMissing` — verify error when file doesn't exist

---

### Task 4: CLI delivery commands for game config

**Subsystem:** `delivery/cli`
**Files:** `backend/internal/delivery/cli/game_config.go` (new)
**Depends on:** Task 2

**What to do:**
Create thin cobra command builders that receive `*app.GameConfigService` and
delegate to it. Each command prints a success message or returns an error.

1. **`CmdCreateGame(svc) *cobra.Command`**
   - Use: `game`
   - Flags: `--path` (string, required) — directory to write `game.json` and `auth.json`
   - Calls `svc.CreateGame(path)`
   - On success: prints `"game created: {path}"`

2. **`CmdAddEmpire(svc) *cobra.Command`**
   - Use: `empire`
   - Flags: `--path` (string, required) — directory containing `game.json` and `auth.json`
   - Flags: `--empire` (int, default 0) — empire number (0 = auto-assign)
   - Calls `svc.AddEmpire(path, empireNo)`
   - On success: prints `"added empire {N}, magic link: {uuid}"`

3. **`CmdRemoveEmpire(svc) *cobra.Command`**
   - Use: `empire`
   - Flags: `--path` (string, required)
   - Flags: `--empire` (int, required)
   - Calls `svc.RemoveEmpire(path, empireNo)`
   - On success: prints `"removed empire {N}"`

4. **`CmdShowMagicLink(svc) *cobra.Command`**
   - Use: `magic-link`
   - Flags: `--path` (string, required)
   - Flags: `--empire` (int, required)
   - Calls `svc.ShowMagicLink(path, empireNo)`
   - On success: prints the URL to stdout (just the URL, nothing else — easy to pipe)

All commands use `RunE` and return errors (not `log.Fatal`). No game logic,
no file I/O, no UUID generation in this layer.

**Acceptance criteria:**
- [ ] `cd backend && go build ./internal/delivery/cli/...` succeeds
- [ ] Handlers are thin — no business logic, no file I/O
- [ ] No imports of `infra` or `runtime`
- [ ] `--path` is required on all commands
- [ ] `--empire` is required on remove and show commands

**Tests to add/update:**
- None — delivery validated via CLI integration

---

### Task 5: Runtime CLI wiring for game config commands

**Subsystem:** `runtime/cli`
**Files:** `backend/internal/runtime/cli/cli.go`
**Depends on:** Tasks 3, 4

**What to do:**
Update `AddCommands` in `backend/internal/runtime/cli/cli.go` to wire the new
`GameConfigService` and attach the new commands.

1. The existing `filestore.Store` instance already satisfies the new
   `GameConfigStore` port (after Task 3), so reuse it.
2. Create `app.GameConfigService{Store: store}`.
3. Add `CmdCreateGame(gameConfigSvc)` to the existing `createCmd` group.
4. Create a `removeCmd` group (`Use: "remove"`, `Short: "remove game objects"`)
   and add `CmdRemoveEmpire(gameConfigSvc)` to it.
5. Add `CmdAddEmpire(gameConfigSvc)` to the existing `createCmd` group.
6. Create a `showCmd` group (or reuse if one exists) and add
   `CmdShowMagicLink(gameConfigSvc)` to it.
7. Add the new command groups to root.

After wiring, the CLI tree should look like:
```
cli create cluster ...
cli create game --path DIR
cli create empire --path DIR [--empire N]
cli remove empire --path DIR --empire N
cli show magic-link --path DIR --empire N
cli show version
```

Note: `cmd/cli/main.go` already has a `cmdShow()` function that creates a
`show` group with `version`. Either pass the show group into `AddCommands`
so it can add `magic-link`, or move the show group into `AddCommands`. Choose
the approach that results in the smallest diff. If you move show into
`AddCommands`, remove the now-unused `cmdShow()` and `cmdShowVersion()` from
`main.go`.

**Acceptance criteria:**
- [ ] `cd backend && go build ./cmd/cli/` succeeds
- [ ] `cli create game --path /tmp/test` creates `game.json` and `auth.json`
- [ ] `cli create empire --path /tmp/test` adds an empire and prints the magic link
- [ ] `cli remove empire --path /tmp/test --empire 1` deactivates the empire
- [ ] `cli show magic-link --path /tmp/test --empire 1` prints the magic link URL
- [ ] All existing tests pass: `cd backend && go test ./...`

**Tests to add/update:**
- None — wiring validated via CLI integration

---

## Task Summary

| Task | Title                                        | Status | Agent/Thread | Notes |
|------|----------------------------------------------|--------|--------------|-------|
| 1    | Domain types for game and auth config        | TODO   |              |       |
| 2    | App-layer port and GameConfigService         | TODO   |              |       |
| 3    | Infra filestore adapter for game/auth config | TODO   |              |       |
| 4    | CLI delivery commands for game config        | TODO   |              |       |
| 5    | Runtime CLI wiring for game config commands  | TODO   |              |       |
