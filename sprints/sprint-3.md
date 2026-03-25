# Sprint 3: Deployment Pipeline

**Pass:** Pass 1
**Goal:** Set up production deployment for the API server and frontend on epimethean.dev.
**Predecessor:** Sprint 2

{{< callout type="warning" >}}
This sprint document was written retroactively after implementation.
Tasks were reconstructed from committed artifacts.
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

This sprint sets up the production deployment pipeline. The API server (`backend/cmd/api/`) is a Go binary that serves the REST API and the frontend static assets. The frontend (`apps/web/`) is a React SPA built with Vite that produces static files in `dist/`.

The target host is `epimethean.dev`. The API server runs as a systemd service under a dedicated user. Caddy or Nginx handles TLS termination and reverse-proxies to the Go server.

**Key files:**
- `backend/DEPLOY.md` — backend deployment notes
- `apps/web/DEPLOY.md` — frontend deployment notes
- `dist/linux/ec-alpha-api.service` — systemd unit file
- `scripts/deploy-backend.sh` — build and deploy script
- `backend/dist/.gitignore` — keeps build output dir tracked but empty

**Build/test commands:**
```bash
cd backend && go build ./...
cd backend && go test ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linux/api ./cmd/api/
```

---

## Tasks

### Task 1: Backend deployment documentation

**Subsystem:** ops
**Files:** `backend/DEPLOY.md`
**Depends on:** None

**What to do:**
Document the production filesystem layout for the API server. Include the system user, directory structure, and ownership for executables (`/opt/ec/alpha`), persistent state (`/var/lib/ec/alpha`), logs (`/var/log/ec/alpha`), and maintenance flags (`/etc/ec/alpha`).

**Acceptance criteria:**
- [x] `backend/DEPLOY.md` exists with directory layout and setup commands
- [x] Commands are commented out (reference, not a runnable script)

**Tests to add/update:**
- None — documentation only

---

### Task 2: Systemd service unit

**Subsystem:** ops
**Files:** `dist/linux/ec-alpha-api.service`
**Depends on:** Task 1

**What to do:**
Create a systemd unit file for the API server. The service should:
- Run as `Type=simple` under a dedicated user
- Set `WorkingDirectory` to the data directory so dotenv files resolve
- Start the binary with `api serve`
- Restart on failure with a 3-second delay
- Refuse to start if `/etc/ec/alpha/MAINTENANCE` exists (`ConditionPathExists`)
- Apply security hardening: `NoNewPrivileges`, `PrivateTmp`, `ProtectSystem=full`, `ProtectHome`, `MemoryDenyWriteExecute`, `RestrictNamespaces`, etc.
- Limit write access to data and log directories via `ReadWritePaths`
- Log to journald with `SyslogIdentifier=ec-alpha`

**Acceptance criteria:**
- [x] Service file passes `systemd-analyze verify` (no syntax errors)
- [x] Maintenance flag prevents startup
- [x] Security hardening directives are present
- [x] Write access is restricted to `/var/lib/ec/alpha` and `/var/log/ec/alpha`

**Tests to add/update:**
- None — validated by systemd on deployment

---

### Task 3: Backend build and deploy script

**Subsystem:** ops
**Files:** `scripts/deploy-backend.sh`, `backend/dist/.gitignore`
**Depends on:** Task 2

**What to do:**
Create a shell script that:
1. Determines the version from `go run ./cmd/api show version`, stripping `+dirty` suffixes
2. Cross-compiles the API binary for `linux/amd64` with `CGO_ENABLED=0`
3. Outputs the binary to `dist/linux/api`
4. Deploys via `rsync` to `epimethean.dev:/var/www/app.epimethean.dev/backend/bin/api-${VERSION}`

Also create `backend/dist/.gitignore` that ignores everything except itself, so the build output directory is tracked but empty.

**Acceptance criteria:**
- [x] Script is executable (`chmod +x`)
- [x] `set -euo pipefail` for safety
- [x] Cross-compilation produces a static linux/amd64 binary
- [x] Version is extracted and stripped of dirty metadata
- [x] `backend/dist/.gitignore` keeps the directory tracked but empty

**Tests to add/update:**
- None — deployment script validated manually

---

### Task 4: Frontend deployment documentation

**Subsystem:** ops
**Files:** `apps/web/DEPLOY.md`
**Depends on:** None

**What to do:**
Document the frontend deployment process: build step (TODO placeholder) and rsync to `epimethean:/var/www/app.epimethean.dev/frontend`.

**Acceptance criteria:**
- [x] `apps/web/DEPLOY.md` exists with rsync command
- [x] Build step is marked as TODO

**Tests to add/update:**
- None — documentation only

---

## Task Summary

| Task | Title                              | Status | Agent/Thread | Notes |
|------|------------------------------------|--------|--------------|-------|
| 1    | Backend deployment documentation   | DONE   |              | Retroactive |
| 2    | Systemd service unit               | DONE   |              | Retroactive |
| 3    | Backend build and deploy script    | DONE   |              | Retroactive |
| 4    | Frontend deployment documentation  | DONE   |              | Retroactive |
