# Roadmap: v0

This document lists the features required to ship v0. It is not a sprint plan — sprints pull from this list. Items are arranged in dependency order so sprint planning can start from the earliest unblocked work. Completed items remain listed when later work depends on them.

---

## MVP Goal

The v0 game engine MVP is a ship that can move. Specifically: a player can build a ship from raw materials, lift it into orbit, and jump it to another orbit in their home system or to another star system.

Everything else in v0 exists to support that goal.

### Critical path to playable v0

1. **Sequence counters** — safe ID generation before ships are created and destroyed
2. **Build Change + Mining Change** — redirect factory and mining output to the needed resource types
3. **Draft** — draft population into specialist roles (ConstructionWorkers, Professionals, etc.)
4. **Pay + Ration** — keep the population productive and fed each turn
5. **Phase 1 & 2 auto-production** — mines produce resources, farms produce food, factories produce units
6. **Transfer** — move manufactured units from ground colony to orbital colony
7. **Assemble** — assemble unit groups from inventory items on a colony or ship
8. **Setup** — create the ship from an assembled set of units
9. **Ship orbit model** — define how a ship tracks its position within a system (orbit number or planet ID, in addition to the system's XYZ coordinates; required before move orders can be implemented)
10. **Move in-system + Move system jump** — jump to another orbit or star system
11. **Turn reports** — text summary per empire of what was produced, what is in inventory, and where ships are; required to close the game loop

---

## Dependency-Ordered Backlog

### 1. Prerequisites already in place

- Magic link → JWT flow (done — Sprint 1)
- Order submission and storage (done — Sprint 1)
- React + Vite player interface (scaffolded — Sprint 2; not feature-complete)
- Cluster generation (done — Sprint 4)
- Dashboard API and cards (done — Sprints 10–11)
- File-backed store for cluster, game state, orders, and reports (in use — evolving through sprints)

### 2. World setup prerequisites

- Homeworld placement and race creation (Sprint 7)
- Empire registration and homeworld assignment (Sprint 7)
- Homeworld template and colony seeding (Sprints 8–9)

### 3. Turn-engine foundation

- **MVP order set + design doc** — define the subset of orders that must work end-to-end in v0. The design doc should cover order text syntax, the `domain.Order` type hierarchy, parse-time vs. execution-time validation, and the turn phase each order maps to.
- **Sequence counters in game file** — add max sequence numbers for deposits, colonies, ships, etc. to `game.json` so ID generation does not rely on `len(slice) + 1`, which breaks if records are ever deleted. Required before ships are created and destroyed in the turn pipeline.
- **Ship orbit/position model** — resolve how ships track location within a system. `Ship.Location` currently identifies the system but not the orbit. Must be decided before move orders can be implemented.
- **Order parsing** — parse submitted order text into structured `domain.Order` values. No execution; parsing and validation only.
- **Turn phase pipeline** — implement the 21-phase sequence of play so the turn engine exists even when many phases are stubs.
- **Order execution** — run turn phases against parsed orders. All phases must be present; some may be stubs in v0.

### 4. MVP economy and production chain

- **Build Change + Mining Change** — redirect factory and mining output to the needed resource types.
- **Draft** — draft population into specialist roles (ConstructionWorkers, Professionals, etc.).
- **Pay + Ration** — keep the population productive and fed each turn.
- **Phase 1 & 2 auto-production** — mines produce resources, farms produce food, and factories produce units automatically each turn.

### 5. MVP ship construction and movement chain

- **Transfer** — move manufactured units from ground colony to orbital colony.
- **Assemble** — assemble unit groups from inventory items on a colony or ship.
- **Set up** — create the ship from an assembled set of units.
- **Move in-system + Move system jump** — jump to another orbit or star system.
- **Name (planet, ship, colony)** — MVP admin support so players can issue meaningful follow-up orders after ships and colonies exist.

### 6. Close the game loop

- **Report generation** — create report files from turn-processing results. The read path (`ListReports`/`GetReport`) already exists but depends on turn processing to produce output.
- **Turn reports** — text summary per empire of what was produced, what is in inventory, and where ships are; required to close the game loop.
- **Empire lifecycle automation** — inactive empires become independent nations with limited v0 moderation: feed the population and return all ships to the nearest colony as soon as possible.

### 7. Frontend rollout, after backend data exists

- **Orders page completion** — add parse and submit actions to the text-based order entry flow.
- **Colonies page enhancement** — expand the summary view once production and inventory data are available.
- **Colony detail page** — inventory, group status, and production.
- **Ships page** — replace placeholder data with real ship data.
- **Ship detail page** — location, inventory, assembled groups, and jump range.
- **Systems page** — visited systems only, sorted by distance.
- **System detail page** — sensor data and nearby systems; optionally in the context of a specific ship.
- **Planets page** — sensor data only; requires a visit.
- **Planet detail page** — sensor data only; requires a visit.
- **Shared UI components** — extract `StatCard`, `DataTable`, and `EmptyState` on second use, triggered by the first interactive page set.
- **Page routing** — reevaluate `useState<Page>` once the MVP page set grows to roughly 13 pages with parameterized detail views.

### 8. Documentation and trigger-based infrastructure

- **Documentation site content** — publish Hugo docs only after behavior is implemented and examples can be checked against real API behavior.
- **SQLite persistence layer** — replace the file-backed store once order parsing and turn-state persistence make whole-file mutation painful. This is a trigger-based change expected after the domain model stabilizes, not a prerequisite for the MVP loop.

---

## Frontend Rollout

- React + Vite player interface (scaffolded — Sprint 2; not feature-complete)
- Dashboard API and cards (done — Sprints 10–11)
- **Shared UI components** — extract reusable `StatCard`, `DataTable`, `EmptyState` components from Sprint 11 patterns. Do this on second use (when the first interactive page is built), not speculatively. The trigger is the order submission UI or colony detail views.
- **Page routing** — the `useState<Page>` pattern in `App.tsx` works through Sprint 11 (~8 pages). The MVP page set is ~13 pages with parameterized detail views; this is the trigger to evaluate `useReducer` or a lightweight router.

### MVP pages

Text-only, turn-based. No maps, no system viewers, no real-time updates.

| Page | Type | Status | Notes |
|------|------|--------|-------|
| Dashboard | summary | done | |
| Colonies | summary | partial | needs enhancement |
| Colony detail | detail | not started | inventory, group status, production |
| Ships | summary | placeholder | needs real data |
| Ship detail | detail | not started | location, inventory, assembled groups, jump range |
| Planets | summary | not started | sensor data only; requires visit |
| Planet detail | detail | not started | sensor data only; requires visit |
| Systems | summary | placeholder | visited systems only, sorted by distance |
| System detail | detail | not started | sensor data + nearby systems; optionally in context of a specific ship (adds that ship's jump range overlay) |
| Orders | entry | partial | needs parse button and submit button; text box only, no syntax UI |
| Reports | summary | done | |
| Report detail | detail | done | |

### Sensor data

Sensor data is available only for systems the player has a ship or colony present in. The data model uses the same attribute set as survey and probe reports; fields not yet implemented are present but marked N/A.

Sensors automatically provide:
- Number of ships in orbit and approximate mass of each
- Number of colonies in orbit, approximate mass of each, and approximate number of production units per colony
- Planet attributes (type, habitability, deposits, etc.) — present in the data model, N/A until Survey/Probe orders are implemented

### Jump range (ship detail only)

Jump range is ship-specific: computed from the ship's HyperEngine units and tech level. The ship detail page lists visited systems this ship can reach. The system detail page shows nearby systems sorted by distance but does not calculate reachability — that is the ship's concern, not the system's.

Note: random cluster generation does not guarantee that any system falls within a given ship's jump range. This is a known v0 limitation, not a UI defect.

---

## Documentation (Hugo site)

- Hugo + Hextra static site at `apps/site/` (scaffolded)
- Must reflect actual system behavior; examples must be checked against real API behavior before publishing.
- Content TBD — see "MVP Documentation" below.

---

## MVP Order Set

The MVP order set is the minimum subset of orders needed for a playable game loop: submit orders → process turn → receive report → repeat. Orders not in the MVP set are rejected at parse time with a clear "not yet implemented" error.

The design doc (to be written before sprint 12) should cover: order text syntax, the `domain.Order` type hierarchy, parse-time vs. execution-time validation, and the turn phase each order maps to.

### Design decisions required before implementation

- **Ship orbit/position model** — `Ship.Location` is currently `Coords` (3D XYZ), which identifies the star system but not the orbit within it. In-system movement requires knowing which orbit (1–10) of which star a ship is currently at. Decision: add `SystemID` + `Orbit int`, or use `PlanetID` for current position, or a combination. Must be resolved before the move orders can be implemented.

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

| Category       | Order                      | Phase | MVP | Notes |
|----------------|----------------------------|-------|-----|-------|
| **Production** | *(none — automatic)*       | 1, 2 | Auto | Mining, farming, and factory output computed automatically each turn |
| **Combat**     | Bombard                    | 3 | Stub | |
| **Combat**     | Invade                     | 3 | Stub | |
| **Combat**     | Raid                       | 3 | Stub | |
| **Combat**     | Support attacker           | 3 | Stub | |
| **Combat**     | Support defender           | 3 | Stub | |
| **Setup**      | Set up (ship/colony)       | 4 | **MVP** | Create new ships and colonies |
| **Assembly**   | Disassemble                | 5 | Stub | |
| **Assembly**   | Build change               | 6 | **MVP** | Redirect factory group output to a different unit type |
| **Assembly**   | Mining change              | 7 | **MVP** | Reassign mining group to a specific deposit |
| **Transfer**   | Transfer                   | 8 | **MVP** | Move units between ships/colonies at same location |
| **Assembly**   | Assemble (factory)         | 9 | **MVP** | Assemble factory units into a factory group |
| **Assembly**   | Assemble (mine)            | 9 | **MVP** | Assemble mine units into a mining group |
| **Assembly**   | Assemble (other)           | 9 | **MVP** | Assemble other unit types (drives, life support, etc.) |
| **Market**     | Buy                        | 10 | Stub | |
| **Market**     | Sell                       | 10 | Stub | |
| **Recon**      | Survey                     | 11 | Stub | |
| **Recon**      | Probe                      | 12 | Stub | |
| **Espionage**  | Check rebels               | 13 | Stub | |
| **Espionage**  | Convert rebels             | 13 | Stub | |
| **Espionage**  | Incite rebels              | 13 | Stub | |
| **Espionage**  | Check for spies            | 13 | Stub | |
| **Espionage**  | Attack spies               | 13 | Stub | |
| **Espionage**  | Gather information         | 13 | Stub | |
| **Movement**   | Move (in-system)           | 14 | **MVP** | Jump to another orbit in same system |
| **Movement**   | Move (system jump)         | 14 | **MVP** | Jump to another star system |
| **Draft**      | Draft                      | 15 | **MVP** | Draft population into specialist roles |
| **Draft**      | Disband                    | 15 | Stub | |
| **Pay/Ration** | Pay                        | 16 | **MVP** | Set wages by population type |
| **Pay/Ration** | Ration                     | 16 | **MVP** | Set food ration percentage |
| **Population** | *(none — automatic)*       | 17, 18, 20 | Auto | Rebellion, rebel growth, and population growth computed automatically |
| **Admin**      | Name (planet)              | 19 | **MVP** | |
| **Admin**      | Name (ship/colony)         | 19 | **MVP** | |
| **Admin**      | Control                    | 19 | Stub | Claim control of a location |
| **Admin**      | Un-control                 | 19 | Stub | Release control of a location |
| **Diplomacy**  | Permission (trade station) | 10 | Stub | |
| **Diplomacy**  | Permission to colonize     | 19 | Stub | |
| **Comms**      | News (market planet)       | 21 | Stub | |
| **Comms**      | News (trade station)       | 21 | Stub | |

### MVP rationale

The MVP goal is a ship that can move. The full dependency chain runs from raw material extraction through manufacturing, assembly, ship construction, and finally jump orders. Every order promoted to MVP is on that critical path:

- **Mining change + Build change** — without these, players cannot redirect their homeworld's production toward ship components; mines default to unspecified deposits and factories default to consumer goods.
- **Draft** — population must be drafted into ConstructionWorkers and Professionals before mines and factories run at useful capacity.
- **Pay + Ration** — colony economics; unpaid or starving populations trigger rebellion, which halts production.
- **Auto phases 1 & 2** — mines, farms, and factories run without orders; output accumulates in colony inventory each turn.
- **Assemble (all variants)** — manufactured units sit in inventory as disassembled items until assembled into groups; drives, life support, and structural components must be assembled before a ship can be set up.
- **Transfer** — units must be moved from ground colony to orbital colony before a ship can be set up in orbit.
- **Setup** — creates the ship from an assembled set of units at an orbital colony.
- **Move (in-system + system jump)** — the stated goal.
- **Name** — players need to label their ships and colonies to issue meaningful orders.

All 21 turn phases must exist in the pipeline. Non-MVP phases either run automatically (production, population, rebellion) or accept no orders and produce no effect (combat, market, espionage, etc.).
