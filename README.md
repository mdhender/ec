# EC — Epimethean Challenge

A turn-based multiplayer strategy game server inspired by classic play-by-mail game rules. Players command space empires across a procedurally generated cluster, submitting orders each turn that are batch-processed by the server.

Code licensed under the [GNU Affero General Public License v3](LICENSE). Rules text licensed under CC BY-NC-ND 4.0.

---

## Repository Layout

```
apps/site/    Hugo + Hextra documentation site
apps/web/     React + Vite + TailwindCSS player interface
backend/      Go API server and CLI (share one SQLite database)
docs/         Architecture and developer guides
ops/          Deployment configuration (systemd, Caddy)
scripts/      Build and deploy helper scripts
archives/     Historical tabletop rules (1980s–1990s)
```

---

## Quick Start

**Backend**

```bash
cd backend
go build ./...
go test ./...
```

Run the API server:

```bash
go run ./cmd/api/ --data-path ./data --jwt-secret <secret>
```

Run the CLI:

```bash
go run ./cmd/cli/ --data-path ./data <command>
```

Configuration is resolved in order: flags → environment variables (`EC_*`) → `.env` file → defaults. See [`backend/cmd/api/README.md`](backend/cmd/api/README.md) and [`backend/cmd/cli/README.md`](backend/cmd/cli/README.md) for full option reference.

**Frontend**

```bash
cd apps/web
npm install
npm run dev       # development server
npm run build     # production build
```

**Documentation site**

```bash
./scripts/hugo-server.sh    # local preview
./scripts/hugo-deploy.sh    # production deploy
```

---

## Architecture

The backend enforces the **SOUSA** layered architecture with strict inward-only dependencies:

```
domain ← app ← infra / delivery ← runtime
```

| Layer | Path | Responsibility |
|---|---|---|
| `domain` | `backend/internal/domain/` | Pure game rules — no framework imports |
| `cerr` | `backend/internal/cerr/` | Canonical sentinel errors |
| `app` | `backend/internal/app/` | Use cases, port interfaces, orchestration |
| `infra` | `backend/internal/infra/` | SQLite, filestore, auth adapters |
| `delivery` | `backend/internal/delivery/` | HTTP (Echo) handlers and CLI commands |
| `runtime` | `backend/internal/runtime/` | Wires concrete types; owns startup |

See [`docs/SOUSA.md`](docs/SOUSA.md) for the canonical rules and [`docs/HUMANS.md`](docs/HUMANS.md) for a plain-language explanation.

---

## Operational Model

The API server and CLI share a single SQLite database. Turn processing is a batch workflow: the API server is stopped (or put into maintenance mode) before the CLI runs turn-processing commands, giving it exclusive write access. This is an intentional design constraint managed at the deployment level, not in application code.

---

## Development Setup

See [`docs/dev-setup.md`](docs/dev-setup.md) for a full walkthrough.

---

## Game Concepts

- **Empires** — Player-controlled civilizations
- **Cluster** — Procedurally generated star systems with stars, planets, deposits, and colonies
- **Ships** — Military and transport vessels operated by empires
- **Orders** — Commands submitted each turn (movement, combat, colonization, etc.)
- **Turns** — Quarterly game phases resolved in batch
- **Reports** — Per-empire output generated after each turn
