# Sprint 1: API Server v0

**Pass:** Pass 1
**Goal:** Deliver a working API server with magic-link authn, JWT authz, file-backed orders/reports, and graceful shutdown.
**Predecessor:** None

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

This sprint builds the EC API server from scratch. The project follows SOUSA layering (see `docs/SOUSA.md`). Dependencies flow inward only: `domain ← app ← infra / delivery ← runtime`.

Echo v5 (`github.com/labstack/echo/v5`) is the HTTP framework. Handlers use the factory pattern returning `func(c *echo.Context) error`. JWT uses `github.com/golang-jwt/jwt/v5` and `github.com/labstack/echo-jwt/v5`.

The dotenv wrapper (`internal/dotfiles`) loads environment variables before Cobra runs. Env vars use the `EC_` prefix (e.g., `EC_JWT_SECRET`). Cobra flags default to env var values via `os.Getenv` in the variable initializers — there is no Viper binding.

Authentication: users present a magic link (UUID) via `POST /api/login/{magicLink}`. The server loads `auth.json` from the data path, compares magic links with `crypto/subtle.ConstantTimeCompare`, and issues a JWT containing the empire number. The same constant-time comparison is used for the shutdown key.

Data layout on disk:
```
{dataPath}/auth.json              — magic link → empire mapping
{dataPath}/{empireNo}/orders.txt  — empire's current orders (text)
{dataPath}/{empireNo}/{turnYear}.{turnQuarter}.json — turn report
```

After completing a task, update sprints/sprint-1.md: check off acceptance criteria (change [ ] to [x]) and change the task status from TODO to DONE in the Task Summary table at the bottom of the file.

**Key files:**
- `backend/cmd/api/main.go` — Cobra commands, env-to-flag wiring
- `backend/internal/cerr/errors.go` — sentinel errors
- `backend/internal/app/` — port interfaces, use-case types
- `backend/internal/infra/auth/` — magic link store, JWT manager
- `backend/internal/infra/filestore/` — orders & reports file adapter
- `backend/internal/delivery/http/` — Echo routes and handlers
- `backend/internal/runtime/server/` — server struct, startup, graceful shutdown
- `backend/data/alpha/auth.json` — sample auth data

**Key types/functions:**
- `cerr.Error` — constant error type (already exists)
- `app.AuthStore`, `app.TokenSigner`, `app.OrderStore`, `app.ReportStore` — port interfaces
- `infra/auth.MagicLinkStore`, `infra/auth.JWTManager` — concrete adapters
- `infra/filestore.Store` — file-based orders/reports adapter
- `delivery/http.AddRoutes` — route wiring function
- `runtime/server.Server` — server struct with graceful shutdown

**Build/test commands:**
```bash
cd backend && go build ./...
cd backend && go test ./...
cd backend && go build ./cmd/api/
```

**Constraints reminder:**
- `domain` must not import `app`, `infra`, `delivery`, or `runtime`
- `app` must not import Echo, JWT concrete packages, or filesystem adapters
- Handler style: `func HandlerName(deps...) func(c *echo.Context) error`
- Use `crypto/subtle.ConstantTimeCompare` for magic link and shutdown key comparison
- Reference `../gemgem` for patterns (JWT manager, shutdown handler, server start)

---

## Tasks

### Task 1: Add sentinel errors to cerr

**Subsystem:** `cerr`
**Files:** `backend/internal/cerr/errors.go`
**Depends on:** None

**What to do:**
Add sentinel error constants to `backend/internal/cerr/errors.go`. These are used by the `app` and `delivery` layers to communicate business errors without coupling to HTTP status codes.

Add these constants:
```go
const (
    ErrUnauthorized     = Error("unauthorized")
    ErrForbidden        = Error("forbidden")
    ErrNotFound         = Error("not found")
    ErrInvalidMagicLink = Error("invalid magic link")
    ErrInvalidToken     = Error("invalid token")
    ErrMissingToken     = Error("missing token")
    ErrInvalidEmpire    = Error("invalid empire number")
)
```

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/cerr/...`
- [ ] All constants are of type `cerr.Error`
- [ ] No imports of outer layers

**Tests to add/update:**
- `TestErrorsImplementError` in `backend/internal/cerr/errors_test.go` — verify each constant satisfies the `error` interface and `.Error()` returns the expected string

---

### Task 2: Define app-layer port interfaces

**Subsystem:** `app`
**Files:** `backend/internal/app/ports.go`
**Depends on:** Task 1

**What to do:**
Create `backend/internal/app/ports.go` defining the port interfaces that `infra` adapters will implement. These are the contracts between the application layer and infrastructure.

```go
package app

import "context"

// AuthStore loads and validates magic links.
type AuthStore interface {
    ValidateMagicLink(ctx context.Context, magicLink string) (empireNo int, ok bool, err error)
}

// TokenSigner issues and validates JWT tokens.
type TokenSigner interface {
    Issue(empireNo int) (token string, err error)
    Validate(token string) (empireNo int, err error)
}

// OrderStore reads and writes empire order files.
type OrderStore interface {
    GetOrders(ctx context.Context, empireNo int) (string, error)
    PutOrders(ctx context.Context, empireNo int, body string) error
}

// ReportStore lists and reads empire turn reports.
type ReportStore interface {
    ListReports(ctx context.Context, empireNo int) ([]ReportMeta, error)
    GetReport(ctx context.Context, empireNo int, turnYear, turnQuarter int) ([]byte, error)
}

// ReportMeta is metadata about a turn report (for listing).
type ReportMeta struct {
    TurnYear    int    `json:"turn_year"`
    TurnQuarter int    `json:"turn_quarter"`
}
```

Only import `context` and standard library. Do not import `cerr` (errors from infra adapters will be wrapped or mapped in delivery).

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/app/...`
- [ ] No imports of `infra`, `delivery`, `runtime`, Echo, or JWT packages
- [ ] All four interfaces are exported

**Tests to add/update:**
- No tests needed — these are interface definitions only

---

### Task 3: JWT manager in infra/auth

**Subsystem:** `infra/auth`
**Files:** `backend/internal/infra/auth/jwt.go`
**Depends on:** Task 2

**What to do:**
Create `backend/internal/infra/auth/jwt.go` implementing a JWT manager that satisfies `app.TokenSigner`. Use `github.com/golang-jwt/jwt/v5` with HMAC-SHA256 signing.

The manager:
- Is constructed with `NewJWTManager(secret string, ttl time.Duration) *JWTManager`
- `Issue(empireNo int)` — creates a JWT with `sub` set to the string representation of empireNo, `iss` = `"ec"`, `aud` = `["ec-web"]`, standard `iat`/`exp`/`nbf` claims. Returns signed token string.
- `Validate(token string)` — parses and validates the token. Returns the empireNo from the `sub` claim. Returns `cerr.ErrInvalidToken` on any validation failure.

Claims struct:
```go
type Claims struct {
    jwt.RegisteredClaims
}
```

Also add a `Middleware() echo.MiddlewareFunc` method that returns `echo-jwt` middleware configured with this manager's key func, following the pattern in `gemgem/internal/jwtmgr/manager.go`.

Also add a `FromContext(c *echo.Context) (empireNo int, ok bool)` helper that extracts the empire number from a validated JWT in the Echo context.

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/infra/auth/...`
- [ ] `JWTManager` satisfies `app.TokenSigner` interface
- [ ] Issued tokens can be validated back to the original empireNo
- [ ] Expired tokens return an error from `Validate`

**Tests to add/update:**
- `TestJWTRoundTrip` in `backend/internal/infra/auth/jwt_test.go` — issue a token and validate it, verify empireNo matches
- `TestJWTExpired` in `backend/internal/infra/auth/jwt_test.go` — issue a token with very short TTL, sleep, verify validation fails

---

### Task 4: Magic link store in infra/auth

**Subsystem:** `infra/auth`
**Files:** `backend/internal/infra/auth/magiclinks.go`
**Depends on:** Task 2

**What to do:**
Create `backend/internal/infra/auth/magiclinks.go` implementing `app.AuthStore`. This adapter loads magic links from a JSON file on disk.

The file format matches `backend/data/alpha/auth.json`:
```json
{
  "magic-links": {
    "81ce2bb6-42fe-49b2-80c5-0558787c8471": {"empire": 1812},
    "37e81785-84ee-4fee-850b-160e373a4539": {"empire": 42}
  }
}
```

Constructor: `NewMagicLinkStore(path string) (*MagicLinkStore, error)` — reads and parses the JSON file at startup. Stores the links in memory.

`ValidateMagicLink(ctx, magicLink)` — uses `crypto/subtle.ConstantTimeCompare` to compare the provided magic link against each stored key. Returns the empire number on match. Returns `0, false, nil` on no match.

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/infra/auth/...`
- [ ] `MagicLinkStore` satisfies `app.AuthStore` interface
- [ ] Valid magic link returns correct empire number
- [ ] Invalid magic link returns `false` without error
- [ ] Invalid JSON file path returns error from constructor

**Tests to add/update:**
- `TestMagicLinkValid` in `backend/internal/infra/auth/magiclinks_test.go` — write a temp JSON file, load it, validate a known link
- `TestMagicLinkInvalid` in `backend/internal/infra/auth/magiclinks_test.go` — validate a link not in the file, verify `ok == false`
- `TestMagicLinkBadFile` in `backend/internal/infra/auth/magiclinks_test.go` — pass a nonexistent path, verify constructor returns error

---

### Task 5: File-based order and report store in infra/filestore

**Subsystem:** `infra/filestore`
**Files:** `backend/internal/infra/filestore/store.go`
**Depends on:** Task 2

**What to do:**
Create `backend/internal/infra/filestore/store.go` implementing both `app.OrderStore` and `app.ReportStore`.

Constructor: `NewStore(dataPath string) *Store` — stores the base data path.

`GetOrders(ctx, empireNo)` — reads `{dataPath}/{empireNo}/orders.txt`. Returns `cerr.ErrNotFound` if file doesn't exist.

`PutOrders(ctx, empireNo, body)` — writes to `{dataPath}/{empireNo}/orders.txt`. Creates the directory if needed (`os.MkdirAll`). Writes atomically (write to temp file, rename).

`ListReports(ctx, empireNo)` — globs `{dataPath}/{empireNo}/*.*.json`, parses filenames as `{turnYear}.{turnQuarter}.json`, returns slice of `app.ReportMeta` sorted by year then quarter.

`GetReport(ctx, empireNo, turnYear, turnQuarter)` — reads `{dataPath}/{empireNo}/{turnYear}.{turnQuarter}.json`. Returns `cerr.ErrNotFound` if file doesn't exist.

Empire numbers must be formatted as integers with no padding (e.g., `42`, not `0042`).

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/infra/filestore/...`
- [ ] `Store` satisfies both `app.OrderStore` and `app.ReportStore`
- [ ] `PutOrders` followed by `GetOrders` round-trips correctly
- [ ] `GetOrders` on missing file returns `cerr.ErrNotFound`
- [ ] `ListReports` parses filenames and returns sorted results
- [ ] `GetReport` on missing file returns `cerr.ErrNotFound`

**Tests to add/update:**
- `TestOrdersRoundTrip` in `backend/internal/infra/filestore/store_test.go` — write orders, read back, verify content
- `TestOrdersNotFound` in `backend/internal/infra/filestore/store_test.go` — read orders for nonexistent empire, verify `cerr.ErrNotFound`
- `TestListReports` in `backend/internal/infra/filestore/store_test.go` — create temp dir with report files, verify listing returns correct metadata sorted
- `TestGetReportNotFound` in `backend/internal/infra/filestore/store_test.go` — read nonexistent report, verify `cerr.ErrNotFound`

---

### Task 6: HTTP handlers in delivery/http

**Subsystem:** `delivery/http`
**Files:** `backend/internal/delivery/http/handlers.go`
**Depends on:** Tasks 2, 3, 4, 5

**What to do:**
Create `backend/internal/delivery/http/handlers.go` with Echo handler factories using the `func(deps...) func(c *echo.Context) error` pattern.

Handlers to implement:

1. `Todo() func(c *echo.Context) error` — returns 501 "not implemented"

2. `GetHealth() func(c *echo.Context) error` — returns 200 `{"ok": true, "time": "..."}` (RFC 3339 UTC)

3. `PostLogin(authStore app.AuthStore, tokenSigner app.TokenSigner) func(c *echo.Context) error` — extracts `magicLink` from URL param, calls `authStore.ValidateMagicLink`, if valid calls `tokenSigner.Issue`, returns 200 `{"access_token": "...", "token_type": "Bearer"}`. Returns 401 on invalid link.

4. `PostLogout() func(c *echo.Context) error` — returns 200 `{"ok": true}`. No-op.

5. `GetOrders(orderStore app.OrderStore) func(c *echo.Context) error` — extracts `empireNo` from URL param and JWT (via `auth.FromContext`), verifies match, calls `orderStore.GetOrders`, returns 200 with text body. Returns 403 if empireNo doesn't match JWT. Returns 404 if not found.

6. `PostOrders(orderStore app.OrderStore) func(c *echo.Context) error` — same authz check, reads request body as text, calls `orderStore.PutOrders`, returns 200 `{"ok": true}`.

7. `GetReports(reportStore app.ReportStore) func(c *echo.Context) error` — same authz check, calls `reportStore.ListReports`, returns 200 JSON array.

8. `GetReport(reportStore app.ReportStore) func(c *echo.Context) error` — same authz check, extracts `turnYear`/`turnQuarter` from URL params, calls `reportStore.GetReport`, returns 200 with JSON body. Returns 404 if not found.

9. `PostShutdown(key string, shutdownCh chan struct{}) func(c *echo.Context) error` — if key is empty, return 501. Otherwise extract `key` from URL param, use `subtle.ConstantTimeCompare`, send to channel on match. Return 200 `{"ok": true}` or 401.

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/delivery/http/...`
- [ ] All handler factories return `func(c *echo.Context) error`
- [ ] No game logic in handlers — just parse, delegate, format
- [ ] No imports of `runtime` or `infra` concrete types

**Tests to add/update:**
- `TestGetHealth` in `backend/internal/delivery/http/handlers_test.go` — verify 200 and JSON shape
- `TestPostShutdownNoKey` in `backend/internal/delivery/http/handlers_test.go` — verify 501 when no key configured
- `TestPostLogout` in `backend/internal/delivery/http/handlers_test.go` — verify 200 and `{"ok": true}`

---

### Task 7: Route wiring in delivery/http

**Subsystem:** `delivery/http`
**Files:** `backend/internal/delivery/http/routes.go`
**Depends on:** Task 6

**What to do:**
Create `backend/internal/delivery/http/routes.go` with an `AddRoutes` function that wires all handlers to an Echo instance.

```go
func AddRoutes(
    e *echo.Echo,
    jwtMiddleware echo.MiddlewareFunc,
    authStore app.AuthStore,
    tokenSigner app.TokenSigner,
    orderStore app.OrderStore,
    reportStore app.ReportStore,
    shutdownKey string,
    shutdownCh chan struct{},
)
```

Route table:
```
GET  /api/health                              → GetHealth()           (public)
POST /api/login/:magicLink                    → PostLogin(...)        (public)
POST /api/logout                              → PostLogout()          (public)
GET  /api/:empireNo/orders                    → GetOrders(...)        (protected)
POST /api/:empireNo/orders                    → PostOrders(...)       (protected)
GET  /api/:empireNo/reports                   → GetReports(...)       (protected)
GET  /api/:empireNo/reports/:turnYear/:turnQuarter → GetReport(...)   (protected)
POST /api/shutdown/:key                       → PostShutdown(...)     (public, self-authed)
```

Protected routes use a group with `jwtMiddleware` applied. The shutdown route is only registered if `shutdownKey != ""`.

**Acceptance criteria:**
- [ ] File compiles: `go build ./internal/delivery/http/...`
- [ ] Public routes have no middleware
- [ ] Protected routes use JWT middleware group
- [ ] Shutdown route only registered when key is non-empty

**Tests to add/update:**
- No unit tests — route wiring is validated by integration/manual testing in Task 9

---

### Task 8: Server struct and graceful shutdown in runtime/server

**Subsystem:** `runtime/server`
**Files:** `backend/internal/runtime/server/server.go`, `backend/internal/runtime/server/options.go`
**Depends on:** Tasks 3, 7

**What to do:**
Create the server package following the pattern in `gemgem/internal/server/server.go`.

**`options.go`** — Define `Option func(*Server) error` and option functions:
- `WithHost(host string)`, `WithPort(port string)`
- `WithShutdownAfter(d time.Duration)` — auto-shutdown timer (0 to disable)
- `WithShutdownKey(key string)` — enables shutdown route
- `WithAuthStore(s app.AuthStore)`, `WithTokenSigner(s app.TokenSigner)`
- `WithOrderStore(s app.OrderStore)`, `WithReportStore(s app.ReportStore)`

**`server.go`** — `Server` struct and methods:

`New(opts ...Option) (*Server, error)` — applies options, validates required fields (authStore, tokenSigner, orderStore, reportStore), initializes shutdown channel.

`Start() error` — performs these steps:
1. Create `echo.New()`, apply request logger middleware
2. Call `delivery/http.AddRoutes(...)` to wire routes
3. Create `http.Server{Addr: host:port, Handler: e}`
4. Start `srv.ListenAndServe()` in a goroutine
5. Set up signal listener (SIGINT, SIGTERM)
6. If `shutdown.after > 0`, start auto-shutdown timer
7. `select` on shutdown channel, signal context, or timer
8. On any trigger: `srv.Shutdown(ctx)` with 5-second grace period
9. Return nil or shutdown error

The shutdown channel is buffered (size 1). The shutdown trigger uses `sync.Once` to prevent double-close. The pattern mirrors `gemgem/internal/server/server.go` lines 137–253.

**Acceptance criteria:**
- [ ] Files compile: `go build ./internal/runtime/server/...`
- [ ] `New` returns error if required stores are nil
- [ ] Signal, timer, and channel shutdown paths are all wired
- [ ] No imports of `domain`

**Tests to add/update:**
- `TestNewMissingDeps` in `backend/internal/runtime/server/server_test.go` — verify `New` returns error when required options are missing
- `TestShutdownChannel` in `backend/internal/runtime/server/server_test.go` — create server, send to shutdown channel, verify Start returns (use a short test timeout)

---

### Task 9: Cobra `serve` command and go.mod updates

**Subsystem:** `cmd/api`
**Files:** `backend/cmd/api/main.go`, `backend/go.mod`, `backend/go.sum`
**Depends on:** Task 8

**What to do:**
Add a `cmdServe()` function to `backend/cmd/api/main.go` and register it with the root command. Follow the gemgem `cmdServe` pattern for env-to-flag defaults.

Flags and env var defaults (dotfiles loads env before Cobra):
```go
host         := os.Getenv("EC_HOST")          // default "localhost" if empty
port         := os.Getenv("EC_PORT")          // default "8080" if empty
dataPath     := os.Getenv("EC_DATA_PATH")     // required, no default
jwtSecret    := os.Getenv("EC_JWT_SECRET")    // required, no default
shutdownKey  := os.Getenv("EC_SHUTDOWN_KEY")  // optional
var timeout  time.Duration                     // --timeout flag, 0 to disable
```

In `RunE`:
1. Validate `dataPath != ""` and `jwtSecret != ""` — return error if empty
2. Validate `dataPath` is a directory
3. Create `auth.NewMagicLinkStore(filepath.Join(dataPath, "auth.json"))`
4. Create `auth.NewJWTManager(jwtSecret, 24*time.Hour)`
5. Create `filestore.NewStore(dataPath)`
6. Build server with options, call `server.Start()`

Run `go get github.com/labstack/echo/v5`, `go get github.com/golang-jwt/jwt/v5`, `go get github.com/labstack/echo-jwt/v5` to add dependencies.

Also wire `cmdServe()` into the root command: `cmdRoot.AddCommand(cmdServe())`.

**Acceptance criteria:**
- [ ] `go build ./cmd/api/` succeeds
- [ ] `./api serve --help` shows all flags
- [ ] `./api serve` without `--data-path` or `--jwt-secret` returns a clear error
- [ ] With valid flags, server starts and responds to `GET /api/health`
- [ ] SIGINT triggers graceful shutdown
- [ ] `--timeout 5s` auto-shuts down after 5 seconds
- [ ] `--shutdown-key` enables shutdown route; `POST /api/shutdown/{key}` with correct key returns 200

**Tests to add/update:**
- No unit tests — this is integration wiring. Validated manually per acceptance criteria.

---

## Task Summary

| Task | Title                                    | Status      | Agent/Thread | Notes |
|------|------------------------------------------|-------------|--------------|-------|
| 1    | Add sentinel errors to cerr              | TODO        |              |       |
| 2    | Define app-layer port interfaces         | TODO        |              |       |
| 3    | JWT manager in infra/auth                | TODO        |              |       |
| 4    | Magic link store in infra/auth           | TODO        |              |       |
| 5    | File-based order/report store            | TODO        |              |       |
| 6    | HTTP handlers in delivery/http           | TODO        |              |       |
| 7    | Route wiring in delivery/http            | TODO        |              |       |
| 8    | Server struct + graceful shutdown        | TODO        |              |       |
| 9    | Cobra serve command + go.mod             | TODO        |              |       |
