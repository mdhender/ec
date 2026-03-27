---
title: Order Language
---
# Order Language

{{< callout type="info" >}}
   This page covers the v0 order language, implemented in Sprint 12. It reflects what the parser accepts today. The grammar will expand in later releases as more orders are implemented.
{{< /callout >}}

This document describes the grammar for v0 order parsing:

- the text format accepted by the v0 parser
- the `domain.Order` hierarchy the parser emits
- the boundary between parse-time validation and execution-time validation
- the phase number assigned to each parsed order

The parser described here is intentionally smaller than the full 1978 rules language.

## Scope

Included in v0 parsing:

- line-oriented parsing of submitted order text
- typed parsing for the MVP order set
- static validation that does not need live game state
- line diagnostics for malformed, invalid, unsupported, or not-yet-implemented input

Excluded from v0 parsing:

- turn execution
- game-state lookups
- inventory, ownership, and reachability checks
- support for every historical punctuation form used in the original manuals

## Canonical Source Format

The v0 parser accepts a scrubbed, whitespace-oriented order language.

### File-level rules

- Input must be UTF-8 text.
- Line endings are normalized to `\n` before parsing.
- Parsing is case-insensitive outside quoted strings.
- Runs of spaces and tabs outside quoted strings are treated as a single separator.
- Blank lines are ignored.
- Top-level parsing is one order per line, except for `setup`, which is a multi-line block.

### Comments

- `//` starts a comment when it appears outside a quoted string.
- A comment runs to end of line.
- Comment text is removed before tokenization.
- `//` inside a quoted string is preserved as part of the quoted string.

Example:

```text
move 77 orbit 6 // move the scout inward
name ship 39 "Slash // Burn"
```

### Quoted strings

- Double quotes delimit a single string field.
- Quoted strings preserve case and spaces.
- Quotes are required for all names. A name without surrounding quotes is a parse error.
- Surrounding quotes are stripped; the domain value contains only the characters between them.
- Escape sequences are not part of v0. A literal `"` inside a name is not supported.
- An unterminated quoted string is a parse diagnostic.

### Numbers

- IDs, group numbers, and deposit numbers must be bare positive integers (no commas).
- Quantities accept embedded commas as thousands separators; commas are removed before parsing.
- Percentages are bare integers in the range `0`–`100`. No `%` suffix.
- Pay rates are non-negative integers in game currency units.

Examples (quantity field):

- `50000`
- `50,000`

Examples (ID field):

- `91`
- `348`

### Deliberate v0 simplifications

- Canonical syntax is whitespace-separated. Commas and trailing periods from the historical manuals are not part of the v0 grammar.
- Canonical command words are lowercase in this document, but the parser matches them case-insensitively.
- The parser accepts a small alias set for historically common spellings, but new examples and tests should use the canonical forms below.

## Primitive Tokens

### IDs and group numbers

| Token | Meaning | Parse-time rule |
|---|---|---|
| `<id>` | ship or colony ID | integer > 0 |
| `<group-id>` | factory or mining group ID | integer > 0 |
| `<deposit-id>` | deposit ID | integer > 0 |

Parse-time does not verify that the referenced object actually exists.

### Coordinates and locations

`<system-coords>` is a three-part coordinate token:

```text
<x>-<y>-<z>
```

Parse-time rules:

- `x`, `y`, and `z` must be integers in `0..30`

`<orbit-ref>` is either:

- an orbit number: `1` through `10`
- a star-qualified orbit: `<star-seq>-<orbit>` where `<star-seq>` is a single ASCII letter and `<orbit>` is `1..10`

Examples:

- `6`
- `c-4`

### Percentages and decimals

`<percent>`:

- bare integer, no `%` suffix
- parse-time range: `0..100`

`<rate>`:

- non-negative integer in game currency units
- parse-time range: `>= 0`

Examples:

- `50`
- `100`
- `0`
- `5`

### Names

`<name>` is a quoted string: the remaining tokens on the line after the target ID, joined by spaces, which must begin and end with `"`.

Parse-time rules:

- quotes are required
- surrounding quotes are stripped before the name is stored in the domain value
- length 1..24 characters (measured after stripping quotes)

## Units and Population Kinds

The parser maps textual unit tokens to `domain.UnitKind`. Tech-level suffixes are not part of the v0 token vocabulary; use the base name only.

### Canonical population tokens

| Canonical token | Maps to |
|---|---|
| `unemployable` | `domain.Unemployables` |
| `unskilled-worker` | `domain.UnskilledWorkers` |
| `professional` | `domain.Professionals` |
| `soldier` | `domain.Soldiers` |
| `spy` | `domain.Spies` |
| `construction-worker` | `domain.ConstructionWorkers` |
| `rebel` | `domain.Rebels` |

Accepted aliases may include obvious plural forms and the short forms already used
in repository docs where they are unambiguous, for example `professionals`, `soldiers`,
`spies`, `construction-workers`, and `pro`.

### Canonical unit tokens

The canonical unit vocabulary follows lower-case, domain-aligned slugs.

Population and commodity units use the base name only — no tech-level suffix:

- `food`
- `consumer-goods`
- `structural`
- `light-structural`
- `research-point`

Equipment units require a `-N` tech-level suffix. The base name alone is rejected:

| Base name | Example with required suffix |
|---|---|
| `factory` | `factory-6` |
| `mine` | `mine-2` |
| `farm` | `farm-1` |
| `hyper-engine` | `hyper-engine-1` |
| `space-drive` | `space-drive-1` |
| `life-support` | `life-support-3` |
| `sensor` | `sensor-2` |
| `automation` | `automation-1` |
| `transport` | `transport-1` |
| `energy-weapon` | `energy-weapon-2` |
| `energy-shield` | `energy-shield-1` |
| `missile-launcher` | `missile-launcher-1` |

Accepted aliases for some base names: `fact` for `factory`, `hype` for `hyper-engine`, `spac` for `space-drive`, `sen` for `sensor`, `cons` for `consumer-goods`.

### Manufacturing targets

`<build-target>` for the `build change` command accepts:

- `<unit-token>`
- `consumer-goods`

The historical `research` and `retool` targets are recognized but treated as
`not_implemented` in v0.

## Command Set

### MVP commands accepted by the v0 parser

| Command | Canonical form | Notes |
|---|---|---|
| Set up | `setup ... end` | only multi-line order; returns `not_implemented` in v0 |
| Build Change | `build change <id> <group-id> <build-target>` | |
| Mining Change | `mining change <id> <group-id> <deposit-id>` | |
| Transfer | `transfer <source-id> <dest-id> <unit-token> <quantity>` | one item per order |
| Assemble (other) | `assemble <id> <unit-token> <quantity>` | generic units |
| Assemble (factory) | `assemble <id> factory <factory-unit> <quantity> <build-target>` | factories with build target |
| Assemble (mine) | `assemble <id> mine <mine-unit> <quantity> <deposit-id>` | mines assigned to a deposit |
| Move (in-system) | `move <id> orbit <orbit-ref>` | explicit destination kind |
| Move (system jump) | `move <id> system <system-coords>` | explicit destination kind |
| Draft | `draft <id> <population-kind> <quantity>` | v0 specialist drafting only |
| Pay | `pay <id> <population-kind> <rate>` | integer rate |
| Ration | `ration <id> <percent>` | bare integer, no `%` |
| Name (ship) | `name ship <id> <name>` | |
| Name (colony) | `name colony <id> <name>` | |
| Name (planet) | `name planet <id> <name>` | numeric planet ID only |

### Setup

Canonical form:

```text
setup ship from <id>
transfer <quantity> <unit-token>
transfer <quantity> <unit-token>
end
```

or

```text
setup colony from <id>
transfer <quantity> <unit-token>
end
```

Rules:

- `setup` is the only multi-line order.
- `ship` and `colony` are the only valid setup kinds.
- `from <id>` names the existing source ship or colony.
- Each body line must begin with `transfer`.
- `end` closes the block.
- The new ship or colony ID is not supplied by the player; it will come from sequence counters during execution.

Accepted alias:

- `set up` may be accepted as an alias for `setup`

Example:

```text
setup ship from 29
transfer 50000 structural
transfer 5 space-drive-1
transfer 5 life-support-1
transfer 5 food
transfer 5 professional
transfer 1 sensor-1
transfer 10000 fuel
transfer 61 hyper-engine-1
end
```

### Build Change

Canonical form:

```text
build change <id> <group-id> <build-target>
```

Examples:

```text
build change 16 8 hyper-engine-1
build change 16 9 consumer-goods
```

Notes:

- `research` and `retool` are recognized but return `not_implemented` in v0.
- Parse-time does not verify that the group exists or belongs to the source colony.

### Mining Change

Canonical form:

```text
mining change <id> <group-id> <deposit-id>
```

Example:

```text
mining change 348 18 92
```

### Transfer

Canonical form:

```text
transfer <source-id> <dest-id> <unit-token> <quantity>
```

Example:

```text
transfer 22 29 spy 10
```

Rules:

- v0 transfer is single-item only; multiple items require multiple orders.
- Parse-time does not verify location, capacity, or inventory.

### Assemble

There are three assemble forms. The second token after the location ID determines which form is in use.

**Other form** — assembles generic units:

```text
assemble <id> <unit-token> <quantity>
```

**Factory form** — assembles factory units and configures their build target:

```text
assemble <id> factory <factory-unit> <quantity> <build-target>
```

**Mine form** — assembles mine units assigned to a specific deposit:

```text
assemble <id> mine <mine-unit> <quantity> <deposit-id>
```

Examples:

```text
assemble 58 missile-launcher-1 6000
assemble 58 missile-launcher-1 6,000
assemble 91 factory factory-6 54000 hyper-engine-1
assemble 91 factory factory-6 54,000 hyper-engine-1
assemble 83 mine mine-2 25680 92
assemble 83 mine mine-2 25,680 92
```

Notes:

- `factory` and `mine` on the second position are keyword discriminators, not unit tokens.
- The factory form requires the factory unit to be a `factory-N` token; any other unit kind is rejected.
- The mine form requires the mine unit to be a `mine-N` token; any other unit kind is rejected.
- Quantities accept comma thousands-separators.

### Move

In-system move:

```text
move <id> orbit <orbit-ref>
```

System jump:

```text
move <id> system <system-coords>
```

Examples:

```text
move 77 orbit 6
move 88 orbit c-4
move 79 system 4-6-19
```

Notes:

- v0 parsing uses explicit destination kinds (`orbit`, `system`) to avoid ambiguity.
- The parsed order preserves symbolic destination intent; it does not commit the execution model to a particular ship-location storage shape.

### Draft

Canonical form:

```text
draft <id> <population-kind> <quantity>
```

Examples:

```text
draft 13 soldier 3600
draft 16 professional 400
draft 16 construction-worker 250
```

v0 draft coverage:

- accepted targets: `professional`, `soldier`, `spy`, `construction-worker`
- recognized but not implemented: `trainee`

### Pay

Canonical form:

```text
pay <id> <population-kind> <rate>
```

Examples:

```text
pay 38 unskilled-worker 1
pay 38 professional 5
pay 38 soldier 4
```

Parse-time rules:

- rate must be a non-negative integer (game currency units per turn)
- allowed population kinds: `unemployable`, `unskilled-worker`, `professional`, `soldier`, `spy`, `construction-worker`

### Ration

Canonical form:

```text
ration <id> <percent>
```

Example:

```text
ration 16 50
```

### Name

Ship naming:

```text
name ship <id> <name>
```

Colony naming:

```text
name colony <id> <name>
```

Planet naming:

```text
name planet <id> <name>
```

Examples:

```text
name ship 39 "Dragonfire"
name colony 7 "Outpost Beta"
name planet 5 "New Terra"
```

## Non-MVP Commands

The parser distinguishes between:

- known commands that are not implemented in v0
- completely unknown input

Known historical commands outside the v0 MVP set return the diagnostic code
`not_implemented`.

This list includes:

- `bombard`
- `invade`
- `raid`
- `support`
- `disassemble`
- `buy`
- `sell`
- `survey`
- `probe`
- `check rebels`
- `convert rebels`
- `incite rebels`
- `check for spies`
- `attack spies`
- `gather information`
- `disband`
- `control`
- `un-control`
- `permission`
- `news`

Anything else that does not match either the MVP set or the known-not-implemented
set returns `unknown_command`.

## Domain Order Hierarchy

The parser emits typed `domain.Order` values. Domain values do not carry parser
artifacts such as comments, raw text, or line numbers.

The domain model uses a small interface plus concrete order structs.

```text
Order
├── SetUpOrder
├── BuildChangeOrder
├── MiningChangeOrder
├── TransferOrder
├── AssembleOrder          (other form)
├── AssembleFactoryOrder   (factory form)
├── AssembleMineOrder      (mine form)
├── MoveOrder
├── DraftOrder
├── PayOrder
├── RationOrder
└── NameOrder
```

### Order interface

```go
type OrderKind int

type Order interface {
    Kind()      OrderKind
    TurnPhase() Phase
    Validate()  error
}
```

### Support types

```go
// MoveDestination holds the raw parsed destination for a move order.
// Full execution-time resolution requires the ship-location model from a later sprint.
type MoveDestination struct {
    Raw string // e.g. "orbit 6" or "system 4-6-19"
}

// NameTargetKind identifies what kind of entity a NameOrder renames.
type NameTargetKind int

const (
    NameTargetPlanet NameTargetKind = iota + 1
    NameTargetShip
    NameTargetColony
)
```

### Concrete order structs

```go
// SetUpOrder — Phase 2
type SetUpOrder struct {
    ColonyID ColonyID
    NewName  string
    NewKind  UnitKind
}

// BuildChangeOrder — Phase 4
type BuildChangeOrder struct {
    ColonyID       ColonyID
    FactoryGroupID FactoryGroupID
    NewUnitKind    UnitKind
}

// MiningChangeOrder — Phase 5
type MiningChangeOrder struct {
    ColonyID      ColonyID
    MiningGroupID MiningGroupID
    DepositID     DepositID
}

// TransferOrder — Phase 8
type TransferOrder struct {
    SourceID ColonyID
    DestID   ColonyID
    UnitKind UnitKind
    Quantity int
}

// AssembleOrder — Phase 9 (other form)
type AssembleOrder struct {
    ColonyID ColonyID
    UnitKind UnitKind
    Quantity int
}

// AssembleFactoryOrder — Phase 9 (factory form)
type AssembleFactoryOrder struct {
    LocationID  ColonyID
    FactoryUnit UnitKind // must be Factory
    FactoryQty  int
    BuildTarget UnitKind
}

// AssembleMineOrder — Phase 9 (mine form)
type AssembleMineOrder struct {
    LocationID ColonyID
    MineUnit   UnitKind // must be Mine
    MineQty    int
    DepositID  DepositID
}

// MoveOrder — Phase 12
type MoveOrder struct {
    ShipID      ShipID
    Destination MoveDestination
}

// DraftOrder — Phase 15
type DraftOrder struct {
    ColonyID ColonyID
    PopKind  UnitKind
    Quantity int
}

// PayOrder — Phase 16
type PayOrder struct {
    ColonyID ColonyID
    PopKind  UnitKind
    Wage     int // non-negative integer; 0 means no pay
}

// RationOrder — Phase 17
type RationOrder struct {
    ColonyID         ColonyID
    RationPercentage int // 0..100
}

// NameOrder — Phase 21
type NameOrder struct {
    TargetKind NameTargetKind
    TargetID   int // PlanetID, ShipID, or ColonyID depending on TargetKind
    NewName    string
}
```

Design notes:

- `MoveOrder` carries a raw destination string (`orbit 6`, `system 4-6-19`). Full parsing into orbit and system types is deferred until the ship-location model is defined in a later sprint.
- `NameOrder` uses an explicit `TargetKind` field so ship, colony, and planet naming are unambiguous without splitting into separate types.
- The domain model preserves execution-relevant data only. Line numbers belong in app-layer diagnostics.

## Parse-Time Validation

Parse-time validation is owned jointly by `infra` and `domain`:

- `infra/ordertext` tokenizes and maps text to candidate values
- `domain` validates static invariants intrinsic to the order itself

Parse-time validation includes:

- command recognition
- correct number of fields for a recognized form
- correct block shape for `setup`
- valid integer, percentage, decimal, coordinate, and orbit syntax
- valid unit and population tokens
- name quoting and 24-character limit
- static numeric ranges such as orbit `1..10`, percentage `0..100`, and coordinates `0..30`

Parse-time validation does not include:

- does the referenced ship, colony, group, or deposit exist?
- is the source at the same location as the target?
- does the source have enough units or population?
- does a move fit current hyper-engine limits?
- is a build-change target operationally legal for the current factory group?
- does a setup order include enough materials to produce a valid ship or colony?

## Execution-Time Validation

Execution-time validation belongs in the turn engine, not in the parser.

Examples by order family:

| Order family | Parse-time | Execution-time |
|---|---|---|
| Build Change | command shape, group ID syntax, target token validity | group exists, group belongs to source colony, target change is legal now |
| Mining Change | command shape, deposit ID syntax | mining group exists, deposit exists, deposit is reachable/controlled |
| Transfer | source/target ID syntax, quantity syntax, unit token validity | source and target co-located, enough inventory, capacity/life-support checks |
| Assemble | variant selection, quantity syntax, unit token validity | enough disassembled items, enough construction workers, destination can host group |
| Set up | block syntax, transfer line syntax | required materials present, source location valid, new entity can be created |
| Move | destination token shape | ship exists, ship is at a valid origin, destination reachable with current engines/fuel |
| Draft | quantity syntax, allowed population token | enough source population, specialist conversion rules satisfied |
| Pay/Ration | decimal/percent parsing, population kind validity | colony economy effects, starvation, rebellion, carry-forward state |
| Name | target token shape, name length | target exists, player is allowed to rename it |

## Diagnostics

The parse result exposes stable diagnostics with:

- `line`
- `code`
- `message`

Stable diagnostic codes:

| Code | Meaning |
|---|---|
| `unknown_command` | the line does not begin with a recognized command form |
| `not_implemented` | the line matches a known historical command that v0 does not support |
| `syntax` | the command is recognized, but the field count or clause layout is wrong |
| `invalid_value` | a field is present but fails static validation |
| `unterminated_quote` | a quoted field is not closed before end of line |
| `unexpected_end` | `end` appeared outside an open `setup` block |

Diagnostic rules:

- diagnostics are reported in input order
- parsing continues after a bad top-level line
- a bad line does not invalidate previously accepted orders
- a malformed `setup` block produces diagnostics and does not emit a `SetupOrder`
- `accepted_count` counts accepted top-level orders, not subordinate `transfer` lines inside `setup`

## Phase Mapping

The parser assigns a phase number to each accepted order so later turn processing
can group by phase without reparsing.

| Order | Phase |
|---|---|
| Set up | 2 |
| Build Change | 4 |
| Mining Change | 5 |
| Transfer | 8 |
| Assemble | 9 |
| Move | 12 |
| Draft | 15 |
| Pay | 16 |
| Ration | 17 |
| Name | 21 |

Parse-time does not require players to submit orders in phase order. Input order is
preserved, but phase assignment is explicit on the typed order values.

## Compatibility Notes

- v0 canonical syntax follows the scrubbed, whitespace-based style already described in `apps/site/content/docs/developers/reference/agent-reference.md`.
- Historical comma-separated examples remain useful as source material, but they are not the grammar target for v0.
- The parser may accept a narrow alias set for convenience, but every new example added to the codebase should use the canonical forms in this document.
