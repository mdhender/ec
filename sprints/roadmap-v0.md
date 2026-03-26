# Roadmap: v0

This document lists the features required to ship v0. It is not a sprint plan — sprints pull from this list. Items are grouped by concern, not by delivery order.

---

## Game Setup

- Cluster generation (done — Sprint 4)
- Homeworld placement and race creation (Sprint 7)
- Empire registration and homeworld assignment (Sprint 7)

---

## Authentication and Sessions

- Magic link → JWT flow (done — Sprint 1)

---

## Empire Lifecycle

- Active empires play normally.
- Inactive empires become **independent nations** and are computer-moderated.
  - v0 moderation is limited to two behaviors:
    1. Feed the population.
    2. Return all ships to the nearest colony as soon as possible.

---

## Order Handling

- Order submission and storage (done — Sprint 1)
- **Order parsing** — parse submitted order text into structured `domain.Order` values. No execution; parsing and validation only.
- **Order execution** — run turn phases. All phases must be present; some may be stubs in v0. The MVP order set (the subset of the full command list that must work end-to-end) is TBD.

---

## Turn Processing

- Turn phase pipeline (not started)
- Report generation (not started)

---

## Persistence

- File-backed store for cluster, game state, orders, and reports (in use — evolving through sprints)
- **Sequence counters in game file** — add max sequence numbers for deposits, colonies, ships, etc. to `game.json` so ID generation does not rely on `len(slice) + 1`, which breaks if records are ever deleted.
- **SQLite persistence layer** — replace the file-backed store with a SQLite database using `modernc.org/sqlite` (CGo-free) and `zombiezen.com/go/sqlite` (query interface). File store serves as the import source once the models stabilize.

---

## Frontend

- React + Vite player interface (scaffolded — Sprint 2; not feature-complete)
- Order submission UI (not started)
- Turn report viewer (not started)
