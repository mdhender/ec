# cmd/cli

The EC command-line interface for turn processing and administration.

## Configuration Priority

Each flag resolves its value using: **flag → env var → .env file → default**.

This is handled by `ff.Parse` with `ff.WithEnvVarPrefix("EC")`. The env var
name is derived automatically from the flag name (e.g., `--data-path` maps to
`EC_DATA_PATH`).

## Environment File (.env)

The CLI loads a single `.env` file from the **current working directory** at
parse time. The file is optional — if it does not exist, parsing continues
without error.

```bash
cd backend
./tmp/cli <command>
```

The `.env` file uses `KEY=VALUE` format with env var names (not flag names):

```env
EC_DATA_PATH=/var/data/ec
```

**Only `.env` is loaded.** The previous multi-file priority chain
(`.env.development.local`, `.env.local`, `.env.{environment}`, `.env`) was
removed in Sprint 6. If you were using those files, consolidate your settings
into `.env` or set the values as environment variables directly.
