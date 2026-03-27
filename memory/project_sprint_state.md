---
name: Sprint state
description: Current sprint completion state and what's next as of 2026-03-26
type: project
---

Sprints 1–11 are complete. Sprint 11 was the frontend dashboard enhancement (data-driven cards, colonies/ships/star list pages).

**Why:** Tracking where we are so future sessions don't re-derive it.

**How to apply:** Sprint 12 planning starts from here. The roadmap (sprints/roadmap-v0.md) is the authoritative source; this is just a quick-start pointer.

## What's done
- Auth (magic link → JWT): Sprint 1
- Cluster generation: Sprint 4
- Homeworld/race/empire setup + colony seeding: Sprints 7–9
- Dashboard API + frontend: Sprints 10–11

## Next up: Sprint 12
Sprint 12 is expected to cover the order domain model: domain.Order type hierarchy for the MVP order set, order text syntax spec, parse-time vs. execution-time validation split. This is the design doc the roadmap flags as "to be written before sprint 12." The ship orbit/position model decision (how Ship tracks intra-system position) must also be resolved in or before this sprint since it blocks move order implementation.

## Pending uncommitted design decisions
- Ship orbit/position model: Ship.Location is Coords (XYZ) for inter-system use, but intra-system movement needs orbit tracking. Options: add SystemID + Orbit int, use PlanetID for current position, or combination.
