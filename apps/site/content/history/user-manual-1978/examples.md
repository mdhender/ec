---
title: Examples
weight: 210
---

## Farm Group


Farm Group 1 contains 125,000 FAM-1 units producing 12,500,000 units of food per year.

Each FRM-1 unit requires 1 PRO and 1 USK labor per turn.
That is a total of 125,000 PRO and 375,000 USK labor per turn for the group.

The FRM-1 unit requires 0.5 FUEL per turn.
That is a total of 62,500 FUEL for the group per turn.

Farms produce 100 units of FOOD per year.
It's undocumented, but one-quarter of the results are available each turn.
That's what I've seen in the turn reports.

The farm group produces 3,125,000 units of FOOD per turn.

## Factory Group


Factory Group 1 contains 250,000 FCT-1 units producing CNGD units.

The group contains more than 50,000 units, so each unit requires 1 PRO and 3 USK labor per turn.
That is a total of 250,000 PRO and 750,000 USK labor.

Each FCT-1 unit requires 0.5 FUEL per turn.
That is a total of 125,000 FUEL.

Each FCT-1 unit can consume up to 20 mass units per year.
That is a total of 5,000,000 mass units per year.

CNGD units require 0.2 METS and 0.4 NMTS per unit, which is a total of 0.6 mass units per unit.
When we divide 20 mass units per year by 0.6 mass units per CNGD unit, we get 33.3333 CNGD units per year.
For CNGD, that works out to approximately 6.6667 METS and 13.3333 NMTS per year.
That totals to 20 mass units per year, which matches the consumption constraint on the factory unit.

The work in progress (WIP) calculations are weird.

In one year, the factory group will produce 250,000 * 33.3333, which is approximately 8,333,333 units per year.

In the first turn (let's call that Turn 1), the factory group will consume 125,000 FUEL, 1,666,666.6667 METS, and 3,333,333.3333 NMTS.
It will produce 8,333,333 units, but these units are not 100% complete.
They are in WIP and are only 25% complete.

In Turn 2, the factory group will consume 125,000 FUEL but no additional METS or NMTS.
It will not produce any new units, but will work on the existing 8,333,333 units.
They are in still WIP and are now 50% complete.

In Turn 3, the factory group will consume 125,000 FUEL but no additional METS or NMTS.
It will not produce any new units, but will work on the existing 8,333,333 units.
They are in still WIP and are new 75% complete.

In Turn 4, the factory group will consume 125,000 FUEL but no additional METS or NMTS.
It will complete the work on 8,333,333 units, but won't deliver any of them to the player until Turn 5.

2,083,334 units.


Factory Group 7 contains 50,000 FCT-1 units producing RSCH units.

12,500 units.

The example in section 6.3:

Factory Group 8 contains 50,000 FCT-1 units and is building MSS-2.

MSS requires 2 x TL METS and 2 x TL NMTS per unit, for a total of 4 METS and 4 NMTS per unit.
Production starts on turn 3; the first missiles will be completed on turn 7.

The player wants to produce 125,000 missiles per year.
Each MSS-2 requires 8 mass units of material per unit.
125,000 times 8 is 1,000,000 mass units.
The group needs 1,000,000 / 20, which is 50,000 FCT-1 units, to support this production rate.

The factory group will consume 25,000 FUEL per turn.

Turn 3, the factory group will consume 25,000 FUEL, 1,000,000 METS, and 1,000,000 NMTS.
This is enough material to produce 125,000 missiles.
It will produce 125,000 missiles, which are WIP and are 25% complete (WIP/25%).

Turn 4, the factory group will consume 25,000 FUEL and no additional METS or NMTS.
It will process 125,000 missiles, changing them from WIP/25% to WIP/50%.

Turn 5, there is a FUEL shortage and the factory group can consume only 12,500 FUEL.
This is enough fuel to process 62,500 missiles, changing them from WIP/50% to WIP/75%.
The remaining 62,500 missiles that were WIP/50% remaining at WIP/50%.
(Work is throttled, not abandoned when there is a FUEL shortage.)

Turn 6, the shortage is relieved and the factory group can consume 25,000 FUEL.
It will process 125,000 missiles.
The 62,500 missiles that were WIP/75% are now WIP/100% (complete and ready to be delivered to the player).
The 62,500 missiles that were WIP/50% are changed to WIP/75%.

Turn 7, the factory group will consume 12,500 FUEL and no additional METS or NMTS.
It will process 62,000 missiles.
The 62,500 which were WIP/100% are now delivered to the player.
The remaining 62,500 missiles that were WIP/75% and are now WIP/100% (complete and ready to be delivered to the player).

Turn 8, the factory group consumes no additional FUEL or METS or NMTS.
The 62,500 which were WIP/100% are now delivered to the player.

Turn 5 there is a shortage of FUEL.
Turn 6, the shortage is relieved.
Turn 7, 62,500 missiles are completed.
Turn 8, 62,500 missiles are completed.

Each missile will "cost" 4 metallic and 4 non-metallic units. This is done on turn 3, and the missiles will be completed on turn 7. A shortage, which allows his groups to function at half their capacity, occurs on turn 5 and is relieved on turn 6. Therefore, the player will receive 62,500 missiles on turn 7 and 62,500 on turn 8, instead of 125,000 on turn 7. (Since the shortage only took place for one quarter, the delay is for one quarter.) A shortage in a colony will affect all factory units, regardless of the unit they are manufacturing. Fuel is allocated to the mines and farms before the factories; fuel shortages usually affect only the factories.

```text
425,000 FUEL used by FCT-1 units in Turn 0.

```
