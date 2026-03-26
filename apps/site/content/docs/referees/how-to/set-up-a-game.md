---
title: Set Up a New Game and Add the First Empire
---

Use these five commands to initialize a game, generate its star cluster, assign a homeworld, register the first empire, and retrieve the player's login link.

## Prerequisites

- The `cli` binary is built and on your path.
- You have a dedicated data directory for this game (e.g. `/srv/games/game1`).

## Steps

### 1. Create the game

```sh
cli create game --data-path /srv/games/game1
```

Creates `game.json` and `auth.json` in the data directory.

### 2. Generate the star cluster

```sh
cli create cluster --data-path /srv/games/game1
```

Writes `cluster.json`.
The cluster is generated with default seeds. To reproduce a specific layout or try a different one, pass `--seed1` and `--seed2`:

```sh
cli create cluster --data-path /srv/games/game1 --seed1 42 --seed2 99
```

If you want to replace an existing cluster file, add `--overwrite`.

### 3. Reserve a homeworld

```sh
cli create homeworld --data-path /srv/games/game1
```

Reads `game.json` and `cluster.json`, selects a homeworld planet automatically, and records the new race entry.

To pin a specific planet instead of letting the command choose:

```sh
cli create homeworld --data-path /srv/games/game1 --planet 17
```

To change the minimum distance from other homeworlds (default is 3):

```sh
cli create homeworld --data-path /srv/games/game1 --min-distance 5
```

### 4. Register the empire

```sh
cli create empire --data-path /srv/games/game1 --name "First Empire"
```

Reads `game.json` and `auth.json`, assigns empire number 1, and writes the empire and its credentials back to both files.
The command prints the empire number, scrubbed name, and magic-link token.

To assign a specific empire number instead of auto-assigning:

```sh
cli create empire --data-path /srv/games/game1 --name "First Empire" --empire 1
```

To bind the empire to a specific homeworld planet rather than the active one:

```sh
cli create empire --data-path /srv/games/game1 --name "First Empire" --homeworld 17
```

### 5. Get the player's login link

```sh
cli show magic-link --data-path /srv/games/game1 --empire 1
```

Reads `auth.json` and prints `?magic=<token>`. Prepend your server's base URL before sending it to the player:

```sh
cli show magic-link --data-path /srv/games/game1 --empire 1 --base-url https://your-server.example.com
```

The full URL is printed directly and is ready to copy.

## What each step produces

| Command | Writes |
|---|---|
| `create game` | `game.json`, `auth.json` |
| `create cluster` | `cluster.json` (reads nothing) |
| `create homeworld` | updates `game.json` |
| `create empire` | updates `game.json`, `auth.json` |
| `show magic-link` | output only — no files written |

Each command reads exactly the files the previous step created, so the order is mandatory.

## Adding more empires

Repeat steps 3–5 for each additional player: reserve a new homeworld, register the empire with a different name, and send the resulting magic link to that player.
