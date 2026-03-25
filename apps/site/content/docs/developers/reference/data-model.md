# Empyrean Challenge — Entities

Extracted from `baseline.adoc` and extended by project-defined rules.
Each entity lists its key fields, constraints, and relationships.

---

## Game

| Field    | Value                               |
|----------|-------------------------------------|
| Cluster  | Exactly one Cluster                 |
| Empires  | Zero or more associated Empires     |

A Game is the top-level game instance.
It contains exactly one Cluster.
It is also associated with Empire entities.
This describes the domain relationship only and does not fix the in-memory ownership or storage layout.

---

## Empire

| Field      | Value                                        |
|------------|----------------------------------------------|
| ID         | Integer, unique, starts at 1                 |
| Name       | String, unique when converted to lowercase   |
| HomeWorld  | Assigned Planet                              |
| Player     | Optional associated Player                   |
| Race       | Derived from HomeWorld                       |

An Empire is the in-game faction that participates in a game.
Empire IDs are assigned in creation order.
Exactly one Empire is created for each player input record.
The Empire name is taken from the corresponding player record.
Each Empire is assigned a home world when it is created.
An Empire belongs to a Game, not to the Cluster as a physical-world entity.
An Empire may exist without a Player.
If a Player quits, the Empire remains in the game and simply has no associated Player.

---

## Player

| Field    | Value                               |
|----------|-------------------------------------|
| ID       | Integer, unique                     |
| Email    | String, unique                      |
| Empire   | Optional associated Empire          |
| Identity | May be addressed by ID or email     |

Player is an out-of-game/controller entity rather than the in-game faction itself.
Player-to-Empire assignment is optional and one-to-one when present.
A Player controls at most one Empire.
An Empire has at most one Player.

---

## Race

| Field      | Value                                     |
|------------|-------------------------------------------|
| Definition | All Empires sharing the same HomeWorld    |
| Kind       | Derived grouping, not separately selected |

Race is determined by origin.
All Empires assigned to the same home world belong to the same Race.

---

## Cluster

A single cluster exists per game (one game per database).

| Field              | Value                                       |
|--------------------|---------------------------------------------|
| Shape              | Cube                                        |
| Side length        | 31                                          |
| Coordinate range   | 0–30 per axis                               |
| Coordinate display | 2-digit zero-padded (e.g. `00`, `07`, `30`) |
| Notation           | `XX-YY-ZZ` (e.g. `28-02-18`)               |
| Stars              | 100                                         |

`prng.IntN(n)` produces `[0, n-1]`. To generate a coordinate in `[0, 30]` use `prng.IntN(31)` (i.e. `IntN(clusterAxisSize)` where `clusterAxisSize = 31`).
The Cluster is the physical world of stars, systems, and planets within a Game.

---

## Star

| Field         | Value                                                                         |
|---------------|-------------------------------------------------------------------------------|
| ID            | Integer, unique, 1–100                                                        |
| X, Y, Z       | Coordinates in range [0, 30] per § 3                                          |
| Sequence      | Integer 1–N assigned within its system per § 5; mapped to A, B, C… in reports |
| Orbits        | 11 (numbered 1 = innermost, 11 = outermost)                                   |
| Display label | `XX-YY-ZZ` for single-star systems; `XX-YY-ZZ/S` for multi-star systems       |

In orders, `XX-YY-ZZ/A` is accepted as an alias for a single-star system but reports never emit the `/A` suffix for single-star systems.

---

## System

| Field         | Value                                                     |
|---------------|-----------------------------------------------------------|
| ID            | Integer, unique, 1–N (N ≤ 100)                            |
| X, Y, Z       | Coordinates shared by all member stars                    |
| Stars         | Ordered slice of member stars, ascending by star ID       |
| Display label | `XX-YY-ZZ` (e.g. `28-02-18`) — used in reports and orders |

Systems are formed after all stars are placed in the cluster (per § 5).
System IDs are assigned 1 to N in the order systems are first encountered iterating stars 1–100.
A system's coordinates equal those of its member stars (all members share the same point).
Stars in the same system occupy each other's orbit 11.
A system with 2 stars is binary; 3 stars, trinary.
Interstellar jumps within a binary/trinary system are treated as 0.2 light years.

---

## Orbit

| Field       | Value                                                             |
|-------------|-------------------------------------------------------------------|
| Number      | 1–11                                                              |
| Occupant    | Empty, a Planet, or another Star in orbit 11 of a multi-star system |
| Orbits 1–10 | May be empty or may contain exactly one planet                    |
| Orbit 11    | Always empty in a single-star system; occupied by another star in a multi-star system |

An orbit is a numbered slot within a star system.
It is not itself a planet.
Planet rules that depend on orbit use the orbit number of the slot the planet occupies.

---

## Planet

| Field                 | Value                                                 |
|-----------------------|-------------------------------------------------------|
| Orbit                 | The orbit number it occupies within its star          |
| Type                  | Terrestrial, Asteroid, Gas Giant                      |
| Habitability          | Integer 0–25                                          |
| Deposits              | Zero or more natural resource deposits                |
| Max resource deposits | 40                                                    |
| Deposit count         | Asteroid `1d40`, Terrestrial `2d20`, Gas Giant `4d10` |
| Home world capacity   | Up to 25 Empires when designated as a home world      |

A planet is the occupant of an orbit, not the orbit itself.
If an orbit is empty, there is no planet for that orbit.
Planet attributes such as habitability, deposits, colonies, and derived limits belong to the planet.

**Type notes:**

- **Gas Giant** — surface colonies on moons only; habitability from 1 to 20
- **Terrestrial** — spherical; not necessarily earth-like; habitability from 0 to 25
- **Asteroid** — represents an entire belt; habitability always 0

**Habitability effects:**

- Max open farms on surface = habitability × 100,000
- Max population without increased rebellion risk = habitability × 10,000,000
- Applies to open colonies only
- A terrestrial planet is habitable when habitability > 0

**Home world rules:**

- An eligible home world must be a terrestrial planet in orbit `3`, `4`, or `5`
- A designated home world may host up to `25` Empires
- Multiple home worlds may exist in the same system
- Home world eligibility is determined per planet, not per system

**Planet assignment algorithm** (per star, orbits 1–10):

```
// Step 1 — Random assignment
for n := 1 to 10:
    r := rnd(1, 100)            // uniform integer, inclusive
    if   r <= 29: planetType[n] = TERRESTRIAL
    elif r <= 34: planetType[n] = ASTEROID
    elif r <= 41: planetType[n] = GAS_GIANT
    else:         planetType[n] = EMPTY

// Step 2 — Cap gas giants at 3; demote excess to ASTEROID, innermost first
for n := 1 to 10:
    if count(planetType[1..10], GAS_GIANT) <= 3: break
    if planetType[n] == GAS_GIANT: planetType[n] = ASTEROID

// Step 3 — Cap asteroids at 2; demote excess to TERRESTRIAL, innermost first
for n := 1 to 10:
    if count(planetType[1..10], ASTEROID) <= 2: break
    if planetType[n] == ASTEROID: planetType[n] = TERRESTRIAL

// Step 4 — Sort non-empty orbits by type (TERRESTRIAL < ASTEROID < GAS_GIANT),
//           preserving which orbit indices are occupied. Empty orbits do not move.
occupied := [n for n in 1..10 where planetType[n] != EMPTY]  // ascending
types    := sort([planetType[n] for n in occupied])          // ascending by type value
for i := 0 to len(occupied)-1:
    planetType[occupied[i]] = types[i]

// Step 5 — Create planets in the occupied orbit slots.
for n := 1 to 10:
    if planetType[n] == EMPTY: continue
    planet[n] = Planet{
        Orbit: n,
        Type:  planetType[n],
    }
```

---

## Deposit

Natural resource deposit located on a planet surface (terrestrial), asteroid, or gas giant moon.

| Field          | Value                                       |
|----------------|---------------------------------------------|
| Max per planet | 40                                          |
| Types          | Gold, Fuel, Metallics, Non-metallics        |
| Discovery      | Survey order (exact) or Probe (approximate) |

Mines may only be assigned to a deposit by a surface colony.

**Quantity by type:**

- Asteroid Gold: 100,000–5,000,000 units
- Terrestrial Gold: 100,000–1,000,000 units
- Fuel: 1,000,000–99,000,000 units
- Metallics: 1,000,000–99,000,000 units
- Non-metallics: 1,000,000–99,000,000 units

**Per-planet caps:**

- Terrestrial: max 1 Gold deposit
- Terrestrial: max 5 Fuel deposits
- Gas Giant: max 2 Fuel deposits

**Cap resolution:**

- If Gold is selected after the Gold cap is reached, convert the deposit to Fuel.
- If Fuel is selected after the Fuel cap is reached, convert the deposit by rolling `1d2`: `1` = Metallics, `2` = Non-metallics.

**Yield by planet type:**

- Asteroid Gold: `1d3`
- Asteroid Fuel: `3d6 - 2`
- Asteroid Metallics: `3d10 - 2`
- Asteroid Non-metallics: `3d10 - 2`
- Gas Giant Fuel: `10d4 - 2`
- Gas Giant Metallics: `10d6`
- Gas Giant Non-metallics: `10d6`
- Terrestrial Gold: `1d3`, or `3d4 - 3` if habitability > 0
- Terrestrial Fuel: `10d4 - 2`, or `10d6` if habitability > 0
- Terrestrial Metallics: `10d6`, or `10d8` if habitability > 0
- Terrestrial Non-metallics: `10d6`, or `10d8` if habitability > 0

---

## Resource

| Name          | Abbreviation | Use                               |
|---------------|--------------|-----------------------------------|
| Gold          | GOLD         | Economic exchange, wages          |
| Fuel          | FUEL         | All production and transportation |
| Metallics     | METS         | Manufacturing                     |
| Non-metallics | NMTS         | Manufacturing                     |

Natural resource units each have a mass of 1 mass unit.

---

## Colony

Up to one colony of each type per player per planet.

| Type     | Surface | Location                              | Life Support | Structural units per mass unit |
|----------|---------|---------------------------------------|--------------|--------------------------------|
| Open     | Yes     | Habitable terrestrial only            | Not required | 1                              |
| Enclosed | Yes     | Uninhabitable terrestrial or asteroid | Required     | 5                              |
| Orbiting | No      | Any planet (in orbit)                 | Required     | 10                             |

Units in storage count as ½ their mass toward structural limits.
Structural mass of the colony/ship itself is not counted.

**Trade Station** — a special orbiting colony with trade as its only function.
Minimum: 3,000 structural units + 500 life support units + 100 professional units.
Charges a 1% commission (paid by seller) on completed trades.

**Home Planet Market** — one per race's home planet; same rules as a trade station, but independent of player control.

---

## Ship

| Field                          | Value           |
|--------------------------------|-----------------|
| Quantity per player per planet | Unlimited       |
| Location                       | Always in orbit |
| Life support                   | Required        |
| Structural units per mass unit | 10              |
| Factories                      | Not allowed     |

---

## Population

All population units represent 100 people. All units have a death rate of 0.0625% per turn (non-combat).

| Type                 | Composition                  | Payment (consumer goods/unit/turn) | Notes                                                                            |
|----------------------|------------------------------|------------------------------------|----------------------------------------------------------------------------------|
| Unemployables        | —                            | 0.00                               | Receives all birth increases (0.25%–2.5% of total pop per quarter)               |
| Unskilled Workers    | —                            | 0.125                              | 2% of unemployables → unskilled when unemployables > 30% of total pop            |
| Professionals        | —                            | 0.375                              | Trained from unskilled (5% of trainees graduate/turn; 1 PRO trains 100 trainees) |
| Soldiers             | —                            | 0.25 (0.005 on ships)              | Drafted from unskilled; may not exceed current soldier count per draft           |
| Spies                | 1 professional + 1 soldier   | 0.375 + 0.25                       | Limited to total soldier + professional count                                    |
| Construction Workers | 1 professional + 1 unskilled | 0.375 + 0.125                      | Limited to total professional + unskilled count                                  |
| Rebels               | (tally only)                 | n/a                                | 10% are militia; rebellion triggers when soldiers ≤ 2× rebel militia             |

**Starvation rule** (when ration < 1/16 food unit per pop unit per turn):

```
S = (M - R) / M
// M = 1/16 (minimum food per pop unit without starvation)
// R = rationed amount (< M)
// S = fraction of colony population that starves
```

---

## Farm

| Field                       | Value                                |
|-----------------------------|--------------------------------------|
| TL range                    | 1–10                                 |
| Annual production (TL 1)    | 100 food units                       |
| Annual production (TL 2–10) | 20 × TL food units                   |
| Mass per unit               | 6 + TL                               |
| Fuel per turn               | 0.5 × TL (TL 1–5); 1 × TL (TL 6–10) |
| Labor per unit              | 3 unskilled workers + 1 professional |
| Max FRM-1 per planet        | habitability × 100,000               |

- TL 1: open-colony farm (surface only)
- TL 2–5: hydroponic; surface colonies in orbits 1–5, or orbiting colonies
- TL 6–10: hydroponic with artificial sunlight; ships or colonies beyond orbit 5

---

## Mine

| Field             | Value                                |
|-------------------|--------------------------------------|
| TL range          | 1–10                                 |
| Annual production | 100 × TL mass units of resource      |
| Mass per unit     | 10 + (2 × TL)                        |
| Fuel per turn     | 0.5 × TL                             |
| Labor per unit    | 3 unskilled workers + 1 professional |
| Placement         | Surface colonies only                |

Mines are organized into **mine groups**; each group mines one deposit.

---

## Factory

| Field                      | Value                                 |
|----------------------------|---------------------------------------|
| TL range                   | 1–10                                  |
| Annual production capacity | 20 × TL mass units of output per unit |
| Mass per unit              | 12 + (2 × TL)                         |
| Fuel per turn              | 0.5 × TL                              |
| Placement                  | Colonies only (not ships)             |

Factories are organized into **factory groups**; each group manufactures one unit type.
Manufacturing cycle: 1 year (4 turns). Resources consumed in turn 1; units delivered after turn 4.

**Factory group labor (unskilled workers and professionals per factory unit):**

| Group size   | Professionals | Unskilled Workers |
|--------------|---------------|-------------------|
| 1–4          | 6             | 18                |
| 5–49         | 5             | 15                |
| 50–499       | 4             | 12                |
| 500–4,999    | 3             | 9                 |
| 5,000–49,999 | 2             | 6                 |
| 50,000+      | 1             | 3                 |

Unskilled workers may be replaced by automation units.

**Manufacturing costs (metallics + non-metallics per unit):**

| Unit              | Metallics | Non-metallics |
|-------------------|-----------|---------------|
| Assault Weapon    | 1 × TL    | 1 × TL        |
| Assault Craft     | 3 × TL    | 2 × TL        |
| Anti-missile      | 2 × TL    | 2 × TL        |
| Automation        | 2 × TL    | 2 × TL        |
| Consumer Goods    | 0.2       | 0.4           |
| Energy Shield     | 25 × TL   | 25 × TL       |
| Energy Weapon     | 5 × TL    | 5 × TL        |
| Factory           | 8 + TL    | 4 + TL        |
| Farm              | 4 + TL    | 2 + TL        |
| Hyper Engine      | 25 × TL   | 20 × TL       |
| Life Support      | 3 × TL    | 5 × TL        |
| Light Structural  | 0.01      | 0.04          |
| Military Robots   | 10 + TL   | 10 + TL       |
| Military Supplies | 0.02      | 0.02          |
| Mine              | 5 + TL    | 5 + TL        |
| Missile           | 2 × TL    | 2 × TL        |
| Missile Launcher  | 15 × TL   | 10 × TL       |
| Sensor            | 999 + TL  | 1999 + TL     |
| Space Drive       | 15 × TL   | 10 × TL       |
| Structural        | 0.1       | 0.4           |
| Transport         | 3 × TL    | 1 × TL        |

**Research:** A factory group ordered to do research produces 1 research point per factory unit × TL per year.

**Research points required for TL advancement:**

| Target TL | Research Points |
|-----------|-----------------|
| 2         | 100,000         |
| 3         | 200,000         |
| 4         | 400,000         |
| 5         | 800,000         |
| 6         | 1,600,000       |
| 7         | 3,200,000       |
| 8         | 6,400,000       |
| 9         | 12,800,000      |
| 10        | 25,600,000      |

---

## Weapons Units

### Assault Weapon

| Field     | Value                                   |
|-----------|-----------------------------------------|
| Mass      | 20                                      |
| Fuel      | 0                                       |
| Operation | 1 soldier unit per assault weapon unit  |
| Notes     | Destroyed when its soldier is destroyed |

### Assault Craft

| Field             | Value                     |
|-------------------|---------------------------|
| Mass              | 5 × TL                    |
| Fuel (normal)     | 0.1 per turn              |
| Fuel (combat)     | 0.01 × TL² per round trip |
| Capacity (combat) | 1 soldier unit            |
| Combat factor     | 10 × TL                   |

### Military Robot

| Field         | Value                                  |
|---------------|----------------------------------------|
| Mass          | (2 × TL) + 20                          |
| Fuel          | 0                                      |
| Replaces      | TL × 2 soldier units                   |
| Combat factor | 2 × TL                                 |
| Notes         | Fights until killed (never surrenders) |

### Missile

| Field                        | Value                                            |
|------------------------------|--------------------------------------------------|
| Mass                         | 4 × TL                                           |
| Damage per hit               | 100 × TL mass units                              |
| Hit formula                  | H = M / D² (M = missiles fired, D = distance)    |
| Launcher accuracy adjustment | Subtract avg launcher TL from D², minimum D² = 1 |

### Missile Launcher

| Field        | Value                                                  |
|--------------|--------------------------------------------------------|
| Mass         | 25 × TL                                                |
| Rate of fire | 1 missile or 1 anti-missile per round                  |
| Notes        | TL affects missile accuracy, not anti-missile accuracy |

### Anti-missile

| Field                | Value                                                              |
|----------------------|--------------------------------------------------------------------|
| Mass                 | 4 × TL                                                             |
| Hit rate vs missiles | (50 + TL × 5)%                                                     |
| Notes                | Launched by missile launchers; does not reduce missile launch rate |

### Energy Weapon

| Field           | Value                                                                                           |
|-----------------|-------------------------------------------------------------------------------------------------|
| Mass            | 10 × TL                                                                                         |
| Fuel            | 4 × TL per combat round                                                                         |
| Energy per beam | 10 × TL                                                                                         |
| Damage formula  | DA = ((F / D) × E) − SH  (F = weapons fired, D = distance, E = energy per weapon, SH = shields) |
| Notes           | Cannot fire surface-colony to surface-colony                                                    |

### Energy Shield

| Field      | Value                          |
|------------|--------------------------------|
| Mass       | 50 × TL                        |
| Fuel       | 10 × TL per combat round       |
| Deflection | 10 × TL energy units per round |

### Military Supplies

| Field       | Value                                    |
|-------------|------------------------------------------|
| Mass        | 0.04 per unit                            |
| Consumption | 1 unit per soldier unit per combat round |

---

## Propulsion Units

### Hyper Engine

| Field               | Value                                                                             |
|---------------------|-----------------------------------------------------------------------------------|
| Mass                | 45 × TL                                                                           |
| Jump range          | TL light years                                                                    |
| Capacity            | 1,000 × TL mass units (excluding engine mass)                                     |
| Fuel per jump       | 40 × distance (light years)                                                       |
| Interplanetary jump | Always 0.1 light years                                                            |
| Interstellar jump   | Ends in orbit 11 of target system                                                 |
| Binary system jump  | 0.2 light years, may target any orbit                                             |
| Notes               | Mixed TL engines use lowest TL; only engines needed to move the load consume fuel |

### Space Drive

| Field          | Value                                                                        |
|----------------|------------------------------------------------------------------------------|
| Mass           | 25 × TL                                                                      |
| Fuel           | TL² per combat round (only during combat)                                    |
| Thrust factor  | TL² × 1,000                                                                  |
| Speed (combat) | Total thrust / ship mass = distance per round                                |
| Notes          | Anti-gravity based; cannot be used for interplanetary or interstellar travel |

---

## Support Units

### Sensor

| Field           | Value                                                                                                 |
|-----------------|-------------------------------------------------------------------------------------------------------|
| Mass            | 2,998 + (2 × TL)                                                                                      |
| Fuel            | TL / 20 per turn                                                                                      |
| Probes per turn | TL                                                                                                    |
| Auto-reports    | Ship count + approx mass, colony count + approx mass + approx production, planet count + type + orbit |
| Approximation   | log₁₀(actual value), truncated to integer                                                             |

### Transport

| Field               | Value                                                      |
|---------------------|------------------------------------------------------------|
| Mass                | 4 × TL                                                     |
| Capacity (transfer) | TL² × 200 mass units per turn                              |
| Fuel (transfer)     | TL² / 10 per turn, proportional to % capacity used         |
| Capacity (combat)   | 3 × TL mass units per round trip                           |
| Fuel (combat)       | 0.01 × TL² per round trip                                  |
| Operation           | 1 professional per 10 transports (transfer only)           |
| Notes               | Will not transfer beyond life support or structural limits |

### Life Support

| Field    | Value                |
|----------|----------------------|
| Mass     | 8 × TL               |
| Fuel     | TL per turn          |
| Supports | TL² population units |

### Structural Unit

| Field  | Value        |
|--------|--------------|
| Mass   | 0.5 per unit |
| Houses | 1 mass unit  |

### Light Structural Unit

| Field         | Value                                                         |
|---------------|---------------------------------------------------------------|
| Mass          | 0.05 per unit                                                 |
| Houses        | 1 mass unit                                                   |
| Manufacturing | Orbiting colonies only                                        |
| Notes         | Can substitute for regular structural units without exception |

### Automation

| Field    | Value                                         |
|----------|-----------------------------------------------|
| Mass     | 4 × TL                                        |
| Fuel     | 0                                             |
| Replaces | TL unskilled worker units per automation unit |
| Use      | Factories, farms, or mines                    |

---

## Consumer Goods and Food

### Consumer Goods

| Field  | Value            |
|--------|------------------|
| Mass   | 0.6 per unit     |
| Source | Factories        |
| Use    | Population wages |

### Food

| Field                | Value                                    |
|----------------------|------------------------------------------|
| Mass                 | 6 per unit                               |
| Consumption          | 0.25 units per population unit per turn  |
| Starvation threshold | < 1/16 unit per population unit per turn |

---

## Spy Unit

Composed unit: 1 soldier + 1 professional.
One function per unit per turn; continues until re-ordered.

| Code | Function                                                                                                   |
|------|------------------------------------------------------------------------------------------------------------|
| A    | Report rebel quantity and type                                                                             |
| B    | Convert 1 rebel unit to loyal population                                                                   |
| C    | Uncover foreign spy units (count + origin)                                                                 |
| D    | Suppress foreign spies (3 attackers destroy 1 defender; 1/6 of defenders destroy attackers; ±50% variance) |
| E    | Incite rebellion (convert 1 loyal unit to rebel)                                                           |
| F    | Obtain secrets (1 line of target's turn report per spy per turn)                                           |

---

## Turn

| Field          | Value                                          |
|----------------|------------------------------------------------|
| Value          | Integer, starts at 0 and increments by 1       |
| Setup turn     | Turn 0                                         |
| Time scale     | 1 turn = 1 quarter of a galactic standard year |
| Report display | `year.quarter`                                 |

Turn `0` is the setup turn and displays as `0.0`.

For report display, turns after setup map to year and quarter as follows:

- Turn `1` = `1.1`
- Turn `2` = `1.2`
- Turn `3` = `1.3`
- Turn `4` = `1.4`
- Turn `5` = `2.1`

The mapping is:

- For turn `0`, display `0.0`
- For turn `n > 0`, display `( ((n - 1) / 4) + 1 ).( ((n - 1) % 4) + 1 )`

---

## Empire Assignment

| Field              | Value                                                                       |
|--------------------|-----------------------------------------------------------------------------|
| Current HomeWorld  | The planet currently being used for Empire assignment                       |
| Capacity per world | 25 Empires                                                                  |
| Initial selection  | Choose the first eligible home world at random                              |
| Replacement rule   | When full, choose a new eligible unused home world at random                |
| Failure mode       | Return an error and terminate setup if no required home world can be chosen |
| Randomness         | Uses the game engine PRNG                                                   |

Empire creation before game start uses a current home world.
The first current home world is chosen at random from all eligible planets.
New Empires are assigned to that home world until it reaches capacity.
When a new home world is needed and none is available, Empire creation returns an error and setup terminates.
Given the same seed and the same player input, Empire creation must produce the same results.

---

## Turn Sequence

Executed in this order each turn:

1. Mining and farming production
2. Manufacturing production
3. Combat
4. Set up orders
5. Dis-assembly orders
6. Build change orders
7. Mining change orders
8. Transfers
9. Assembly orders
10. Market and trade station activity
11. Surveys
12. Probe and sensor reports
13. Espionage
14. Ship movement
15. Draft orders
16. Pay and ration orders
17. Rebellion
18. Rebel increases
19. Naming and control orders
20. Population increases
21. News service reports

---

## General Rules

- **Mass unit** = 17,000 lbs (≈ 100 people)
- **All fractions truncated** (no rounding)
- **1 turn = 1 quarter of a galactic standard year**
- **Operational units** must be assembled to function and dis-assembled for transfer. Dis-assembly causes a 10% loss (except spy and construction units). Units requiring assembly: space drives, sensors, automation, life support, energy weapons, energy shields, mines, factories, farms, hyper engines, structural units, light structural units, missile launchers.
- **Victory**: control ≥ 100 planets with no other player controlling ≥ 50 planets, for 4 consecutive turns.
