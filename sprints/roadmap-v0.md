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

The MVP order set is the minimum subset of orders needed for a playable game loop: submit orders → process turn → receive report → repeat. Orders not in the MVP set are rejected at parse time with a clear "not yet implemented" error.

The design doc (to be written before sprint 12) should cover: order text syntax, the `domain.Order` type hierarchy, parse-time vs. execution-time validation, and the turn phase each order maps to.

### Turn phases (1978 Sequence of Play)

| # | Phase | Description |
|---|-------|-------------|
| 1 | Mining/farming production | Calculate resource and food output |
| 2 | Manufacturing production | Calculate factory output |
| 3 | Combat | Resolve bombard, invade, raid, support |
| 4 | Set up | Create new ships/colonies |
| 5 | Disassembly | Disassemble units |
| 6 | Build change | Reassign factory production targets |
| 7 | Mining change | Reassign mining groups to deposits |
| 8 | Transfers | Move units between ships/colonies |
| 9 | Assembly | Assemble units (factories, mines, etc.) |
| 10 | Market/trade | Buy/sell on market planets and trade stations |
| 11 | Surveys | Survey local system |
| 12 | Probes/sensors | Compile probe and sensor reports |
| 13 | Espionage | Spy operations |
| 14 | Ship movement | Execute jump orders |
| 15 | Draft | Draft population into specialist roles |
| 16 | Pay/ration | Set wages and food rations |
| 17 | Rebellion | Resolve rebellions |
| 18 | Rebel increases | Calculate rebel growth |
| 19 | Naming/control | Process name and control orders |
| 20 | Population | Calculate population growth |
| 21 | News | Compile news service reports |

### Order table

"MVP" = must work end-to-end in v0. "Stub" = phase exists but order is rejected as not-yet-implemented. "Auto" = no player order; the phase runs automatically.

| Category | Order | Phase | MVP | Notes |
|----------|-------|-------|-----|-------|
| **Production** | *(none — automatic)* | 1, 2 | Auto | Mining, farming, and factory output computed automatically each turn |
| **Combat** | Bombard | 3 | Stub | |
| **Combat** | Invade | 3 | Stub | |
| **Combat** | Raid | 3 | Stub | |
| **Combat** | Support attacker | 3 | Stub | |
| **Combat** | Support defender | 3 | Stub | |
| **Setup** | Set up (ship/colony) | 4 | **MVP** | Create new ships and colonies |
| **Assembly** | Disassemble | 5 | Stub | |
| **Assembly** | Build change | 6 | Stub | Reassign factory group output |
| **Assembly** | Mining change | 7 | Stub | Reassign mining group to new deposit |
| **Transfer** | Transfer | 8 | Stub | |
| **Assembly** | Assemble (factory) | 9 | Stub | |
| **Assembly** | Assemble (mine) | 9 | Stub | |
| **Assembly** | Assemble (other) | 9 | Stub | |
| **Market** | Buy | 10 | Stub | |
| **Market** | Sell | 10 | Stub | |
| **Recon** | Survey | 11 | Stub | |
| **Recon** | Probe | 12 | Stub | |
| **Espionage** | Check rebels | 13 | Stub | |
| **Espionage** | Convert rebels | 13 | Stub | |
| **Espionage** | Incite rebels | 13 | Stub | |
| **Espionage** | Check for spies | 13 | Stub | |
| **Espionage** | Attack spies | 13 | Stub | |
| **Espionage** | Gather information | 13 | Stub | |
| **Movement** | Move (in-system) | 14 | **MVP** | Jump to another orbit in same system |
| **Movement** | Move (system jump) | 14 | **MVP** | Jump to another star system |
| **Draft** | Draft | 15 | Stub | |
| **Draft** | Disband | 15 | Stub | |
| **Pay/Ration** | Pay | 16 | **MVP** | Set wages by population type |
| **Pay/Ration** | Ration | 16 | **MVP** | Set food ration percentage |
| **Population** | *(none — automatic)* | 17, 18, 20 | Auto | Rebellion, rebel growth, and population growth computed automatically |
| **Admin** | Name (planet) | 19 | **MVP** | |
| **Admin** | Name (ship/colony) | 19 | **MVP** | |
| **Admin** | Control | 19 | Stub | Claim control of a location |
| **Admin** | Un-control | 19 | Stub | Release control of a location |
| **Diplomacy** | Permission (trade station) | 10 | Stub | |
| **Diplomacy** | Permission to colonize | 19 | Stub | |
| **Comms** | News (market planet) | 21 | Stub | |
| **Comms** | News (trade station) | 21 | Stub | |

### MVP rationale

The MVP set (setup, jump, pay, ration, naming) gives players the core loop: manage colony economics (pay/ration), move ships, establish new ships/colonies, and label things. All 21 turn phases must exist in the pipeline — non-MVP phases either run automatically (production, population, rebellion) or accept no orders and produce no effect (combat, market, espionage, etc.).
