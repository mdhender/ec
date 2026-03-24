# EC Dev Environment Setup

## Prerequisites

- **Go 1.25+** — backend API server
- **Node.js / npm** — frontend dev server
- **Caddy 2** — local TLS reverse proxy (`brew install caddy`)
- **Overmind** — process manager (`brew install overmind`)
- **Air** — Go live-reload (`go install github.com/air-verse/air@latest`)

## Local TLS Proxy (Caddy)

Caddy terminates TLS using its built-in CA and reverse-proxies to the backend and frontend dev servers. The single entry point is:

```
https://ec.localhost:8443/
```

All `*.localhost` domains resolve to `127.0.0.1` automatically in modern browsers and operating systems — no `/etc/hosts` entry is needed.

### Routing

| Path            | Target           | Description             |
|-----------------|------------------|-------------------------|
| `/api/*`        | `localhost:3000` | Go backend (Echo)       |
| Everything else | `localhost:5173` | Vite dev server (React) |

### Caddyfile

The project's block lives in the system-wide Caddyfile at `/opt/homebrew/etc/Caddyfile`. Key settings:

```caddyfile
{
    http_port 8080
    https_port 8443
}

ec.localhost:8443 {
    tls internal
    encode zstd gzip

    log {
        output file /opt/homebrew/var/log/ec.dev.access.log
        format console
    }

    @api path /api/*
    handle @api {
        reverse_proxy localhost:3000
    }

    handle {
        reverse_proxy localhost:5173
    }

    header {
        Strict-Transport-Security "max-age=0"
    }
}
```

After editing the Caddyfile, reload with:

```sh
caddy reload --config /opt/homebrew/etc/Caddyfile
# or simply:
brew restart caddy
```

> **Note:** Caddy's own `http_port` is `8080`, so the Go backend must listen on a different port (currently `3000`).

## Running the Stack

A `Procfile.dev` in the repo root starts both servers via Overmind:

```
backend:  cd backend && air
frontend: cd apps/web && npm run dev
```

Start everything:

```sh
# Make sure Caddy is running first
brew services start caddy

# Then start backend + frontend
overmind start -f Procfile.dev
```

Stop with `overmind stop` or `Ctrl-C`.

### Ports Summary

| Service       | Port | Notes                                  |
|---------------|------|----------------------------------------|
| Caddy HTTPS   | 8443 | Entry point — use this in the browser  |
| Caddy HTTP    | 8080 | Reserved by Caddy (global `http_port`) |
| Go backend    | 3000 | Default; override with `-addr` flag    |
| Vite frontend | 5173 | Default Vite dev server port           |

## Trusting the Local CA

On first run Caddy generates a root CA certificate. Most browsers trust it automatically. If you see TLS warnings:

```sh
caddy trust
```

This installs the Caddy root CA into the system trust store (may prompt for your password).
