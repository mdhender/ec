# cmd/api

The EC API server.

## Environment Files (dotenv)

The server loads environment variables from dotenv files **relative to the
current working directory** before Cobra parses flags. Run the binary from the
directory that contains your `.env` files (typically `backend/`):

```bash
cd backend
./tmp/api serve
```

If you run from a different directory, the dotenv files will not be found and
only explicit env vars or flags will take effect.

See `internal/dotfiles/dotfiles.go` for the full priority order.

## Configuration Priority

Each flag resolves its value using: **flag → env var → default**.

| Flag | Env Var | Default |
|---|---|---|
| `--host` | `EC_HOST` | `localhost` |
| `--port` | `EC_PORT` | `8080` |
| `--data-path` | `EC_DATA_PATH` | *(required)* |
| `--jwt-secret` | `EC_JWT_SECRET` | *(required)* |
| `--shutdown-key` | `EC_SHUTDOWN_KEY` | *(empty — disables endpoint)* |
| `--timeout` | `EC_TIMEOUT` | `0` *(disabled)* |
