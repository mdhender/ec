# cmd/cli

The EC command-line interface for turn processing and administration.

## Environment Files (dotenv)

The CLI loads environment variables from dotenv files **relative to the current
working directory** before Cobra parses flags. Run the binary from the directory
that contains your `.env` files (typically `backend/`):

```bash
cd backend
./tmp/cli <command>
```

If you run from a different directory, the dotenv files will not be found and
only explicit env vars or flags will take effect.

See `internal/dotfiles/dotfiles.go` for the full priority order.
