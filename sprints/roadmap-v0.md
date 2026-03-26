# Roadmap: v0

This document lists the features required to ship v0. It is not a sprint plan — sprints pull from this list. Items are grouped by concern, not by delivery order.

---

## Game Setup

- Cluster generation (done — Sprint 4)
- Homeworld placement and race creation (Sprint 7)
- Empire registration and homeworld assignment (Sprint 7)
- Homeworld template and colony seeding (Sprints 8–9)

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
- **MVP order set** — the subset of orders that must work end-to-end for v0. See the "MVP Order Set" section below for the current list.
- **Order parsing** — parse submitted order text into structured `domain.Order` values. No execution; parsing and validation only.
- **Order execution** — run turn phases. All phases must be present; some may be stubs in v0.

---

## Turn Processing

- Turn phase pipeline (not started)
- Report generation (not started) — nothing currently *creates* report files; the read path (`ListReports`/`GetReport`) exists but depends on the turn pipeline to produce output.

---

## Persistence

- File-backed store for cluster, game state, orders, and reports (in use — evolving through sprints)
- **Sequence counters in game file** — add max sequence numbers for deposits, colonies, ships, etc. to `game.json` so ID generation does not rely on `len(slice) + 1`, which breaks if records are ever deleted.
- **SQLite persistence layer** — replace the file-backed store with a SQLite database using `modernc.org/sqlite` (CGo-free) and `zombiezen.com/go/sqlite` (query interface). File store serves as the import source once the models stabilize. **Trigger:** SQLite is needed when order parsing and turn-state persistence make the "read whole file, mutate, write whole file" model painful. Expected around sprint 13–14, after the domain model stabilizes.

---

## Frontend

- React + Vite player interface (scaffolded — Sprint 2; not feature-complete)
- Dashboard API and cards (Sprints 10–11)
- **Shared UI components** — extract reusable `StatCard`, `DataTable`, `EmptyState` components from Sprint 11 patterns. Do this on second use (when the first interactive page is built), not speculatively. The trigger is the order submission UI or colony detail views.
- **Page routing** — the `useState<Page>` pattern in `App.tsx` works through Sprint 11 (~8 pages). Evaluate `useReducer` or a lightweight router when colony/ship detail pages are added.
- Order submission UI (not started)
- Turn report viewer (not started)

---

## MVP Order Set

> **Status:** TBD — define before writing sprint 12.

The MVP order set is the minimum subset of orders needed for a playable game loop: submit orders → process turn → receive report → repeat. Orders not in the MVP set are rejected at parse time with a clear "not yet implemented" error.

Candidates (to be confirmed):

| Order | Purpose | Notes |
|-------|---------|-------|
| *TBD* | *TBD* | *TBD* |

The order set must be defined before speccing the order-parsing sprint. The design doc should cover: order text syntax, the `domain.Order` type hierarchy, parse-time vs. execution-time validation, and the turn phase each order maps to.
