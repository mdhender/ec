---
name: v0 MVP Goals
description: Agreed MVP targets for game engine, frontend, and docs as of sprint planning session (2026-03-26)
type: project
---

The v0 MVP goal is a ship that can move — specifically, a player can build a ship from raw materials, lift it into orbit, and jump to another orbit or star system.

**Why:** This is the minimum playable loop that proves the game engine works end-to-end.

**How to apply:** All sprint scope decisions for the game engine should be evaluated against this goal. If a feature is not on the critical path to a moving ship, it is a stub in v0.

## Game engine dependency chain (ordered)
1. Sequence counters — safe IDs before ships are created/destroyed
2. Build Change + Mining Change — redirect factory and mining output
3. Draft — population into specialist roles
4. Pay + Ration — keep colony productive
5. Phase 1 & 2 auto-production (automatic)
6. Transfer — move units from ground to orbital colony
7. Assemble — assemble unit groups from inventory
8. Setup — create ship from assembled units
9. Ship orbit model — design decision: how does Ship track intra-system position (SystemID+Orbit, or PlanetID)?
10. Move in-system + Move system jump
11. Turn reports — text summary per empire; closes the game loop

## Frontend MVP (~13 pages, text-only, turn-based)
- Dashboard (done)
- Colonies summary (partial), Colony detail (not started)
- Ships summary (placeholder), Ship detail (not started)
- Planets summary (not started), Planet detail (not started)
- Systems summary (placeholder), System detail (not started)
- Orders entry (partial — needs parse + submit buttons, text box only)
- Reports summary (done), Report detail (done)

Sensor data requires a ship or colony present in the system. Uses survey/probe attribute structure; unimplemented fields are N/A.
Jump range is ship-specific (Ship detail page only). System detail page lists nearby systems by distance but does not calculate reachability.
useState<Page> pattern is past its limit at this page count — router evaluation is explicitly triggered.

## Documentation
Player docs should derive from the frontend pages. Key docs: order syntax reference (same artifact as the parser spec), how to read a turn report, colony economics, ship building, sensor data.
The Hugo site at apps/site/ is player-facing for v0; operator/admin docs are out of scope.
