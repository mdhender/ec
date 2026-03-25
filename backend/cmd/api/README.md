# cmd/api

The EC API server.

## Configuration Priority

Each flag resolves its value using: **flag → env var → .env file → default**.

This is handled by `ff.Parse` with `ff.WithEnvVarPrefix("EC")`. The env var
name is derived automatically from the flag name (e.g., `--data-path` maps to
`EC_DATA_PATH`).

| Flag | Env Var | Default |
|---|---|---|
| `--host` | `EC_HOST` | `localhost` |
| `--port` | `EC_PORT` | `8080` |
| `--data-path` | `EC_DATA_PATH` | *(required)* |
| `--jwt-secret` | `EC_JWT_SECRET` | *(required)* |
| `--shutdown-key` | `EC_SHUTDOWN_KEY` | *(empty — disables endpoint)* |
| `--timeout` | `EC_TIMEOUT` | `0` *(disabled)* |

## Environment File (.env)

The server loads a single `.env` file from the **current working directory** at
parse time. The file is optional — if it does not exist, parsing continues
without error.

```bash
cd backend
./tmp/api serve
```

The `.env` file uses `KEY=VALUE` format with env var names (not flag names):

```env
EC_DATA_PATH=/var/data/ec
EC_JWT_SECRET=supersecret
```

**Only `.env` is loaded.** The previous multi-file priority chain
(`.env.development.local`, `.env.local`, `.env.{environment}`, `.env`) was
removed in Sprint 6. If you were using those files, consolidate your settings
into `.env` or set the values as environment variables directly.
