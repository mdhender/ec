# Order Language v0

This document defines the internal reference grammar for Sprint 12 order parsing.
It is the source of truth for:

- the text format accepted by the v0 parser
- the `domain.Order` hierarchy the parser emits
- the boundary between parse-time validation and execution-time validation
- the phase number assigned to each parsed order

This is a v0 design document, not a historical compatibility document. The parser
described here is intentionally smaller than the full 1978 rules language.

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
77 move orbit 6 // move the scout inward
39 name "Slash // Burn"
```

### Quoted strings

- Double quotes delimit a single string field.
- Quoted strings preserve case and spaces.
- Quotes are required for names.
- Escape sequences are not part of v0. A literal `"` inside a name is not supported.
- An unterminated quoted string is a parse diagnostic.

### Numbers

- Positive integers may contain embedded commas for readability.
- Commas are removed before integer parsing.
- IDs, group numbers, deposit numbers, and quantities must be positive integers.
- Percentages are integers followed immediately by `%`.
- Pay rates are decimal literals with up to three fractional digits.

Examples:

- `50000`
- `50,000`
- `75%`
- `0.125`

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

`<planet-ref>` is:

```text
<system-coords>/<orbit-ref>
```

Examples:

- `5-12-38/2`
- `5-12-38/c-4`

### Percentages and decimals

`<percent>`:

- integer followed by `%`
- parse-time range: `0..100`

`<rate>`:

- decimal literal with up to three fractional digits
- parse-time range: `>= 0`
- domain representation should be fixed-point thousandths, not `float64`

Examples:

- `50%`
- `100%`
- `0.125`
- `1.600`

### Names

`<name>` is a quoted string.

Parse-time rules:

- quotes required
- length 1..24 characters after stripping the outer quotes
- surrounding quotes are not preserved in the domain value

## Units and Population Kinds

The parser should map textual unit tokens to `domain.UnitKind` plus an optional
`domain.TechLevel`.

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

Examples without tech level:

- `food`
- `fuel`
- `gold`
- `metallics`
- `non-metallics`
- `consumer-goods`
- `structural`
- `light-structural`
- `research-point`

Examples with required tech level suffix:

- `factory-6`
- `mine-2`
- `farm-1`
- `hyper-engine-1`
- `space-drive-1`
- `life-support-1`
- `sensor-1`
- `automation-3`
- `transport-2`
- `energy-weapon-4`
- `energy-shield-4`
- `missile-launcher-1`

The parser may accept historically common aliases where they are already present in
repository documentation, such as:

- `fact-6` for `factory-6`
- `hype-1` for `hyper-engine-1`
- `spac-1` for `space-drive-1`
- `sen-1` for `sensor-1`
- `cons` for `consumer-goods`

Alias acceptance exists to reduce friction when players copy examples from older docs.
Canonical docs and tests should use the full tokens.

### Manufacturing targets

`<build-target>` and the factory form of `assemble` use one of these forms:

- `<unit-token>`
- `consumer-goods`

The historical `research` and `retool` targets are recognized but treated as
`not_implemented` in v0.

## Command Set

### MVP commands accepted by the v0 parser

| Command | Canonical form | Notes |
|---|---|---|
| Set up | `setup ... end` | only multi-line order |
| Build Change | `<id> build change <group-id> <build-target>` | historical alias `build-change` not required |
| Mining Change | `<id> mining change <group-id> <deposit-id>` | historical alias `mining` may be accepted |
| Transfer | `<id> transfer <quantity> <unit-token> <id>` | one item per order |
| Assemble (factory) | `<id> assemble <quantity> <factory-unit> <build-target>` | determined by first unit token |
| Assemble (mine) | `<id> assemble <quantity> <mine-unit> <deposit-id>` | determined by first unit token |
| Assemble (other) | `<id> assemble <quantity> <unit-token>` | all other assembly |
| Move (in-system) | `<id> move orbit <orbit-ref>` | explicit destination kind |
| Move (system jump) | `<id> move system <system-coords>` | explicit destination kind |
| Draft | `<id> draft <quantity> <population-kind>` | v0 specialist drafting only |
| Pay | `<id> pay <rate> <population-kind>` | fixed-point decimal |
| Ration | `<id> ration <percent>` | integer percentage |
| Name (ship/colony) | `<id> name <name>` | quoted name required |
| Name (planet) | `<planet-ref> name <name>` | quoted name required |

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
<id> build change <group-id> <build-target>
```

Examples:

```text
16 build change 8 hyper-engine-1
16 build change 9 consumer-goods
```

Notes:

- `research` and `retool` are recognized but return `not_implemented` in v0.
- Parse-time does not verify that the group exists or belongs to the source colony.

### Mining Change

Canonical form:

```text
<id> mining change <group-id> <deposit-id>
```

Example:

```text
348 mining change 18 92
```

Accepted alias:

- `mining` may be accepted as a historical alias for `mining change`

### Transfer

Canonical form:

```text
<id> transfer <quantity> <unit-token> <id>
```

Example:

```text
22 transfer 10 spy 29
```

Rules:

- v0 transfer is single-item only; multiple items require multiple orders.
- Parse-time does not verify location, capacity, or inventory.

### Assemble

Canonical forms:

Factory assembly:

```text
<id> assemble <quantity> <factory-unit> <build-target>
```

Mine assembly:

```text
<id> assemble <quantity> <mine-unit> <deposit-id>
```

Other assembly:

```text
<id> assemble <quantity> <unit-token>
```

Examples:

```text
91 assemble 54000 factory-6 consumer-goods
83 assemble 25680 mine-2 148
58 assemble 6000 missile-launcher-1
```

Variant selection rules:

- if the third field is a factory unit token, parse as factory assembly
- if the third field is a mine unit token, parse as mine assembly
- otherwise parse as other assembly

### Move

In-system move:

```text
<id> move orbit <orbit-ref>
```

System jump:

```text
<id> move system <system-coords>
```

Examples:

```text
77 move orbit 6
88 move orbit c-4
79 move system 4-6-19
```

Notes:

- v0 parsing uses explicit destination kinds (`orbit`, `system`) to avoid ambiguity.
- The parsed order preserves symbolic destination intent; it does not commit the execution model to a particular ship-location storage shape.

### Draft

Canonical form:

```text
<id> draft <quantity> <population-kind>
```

Examples:

```text
13 draft 3600 soldier
16 draft 400 professional
16 draft 250 construction-worker
```

v0 draft coverage:

- accepted targets: `professional`, `soldier`, `spy`, `construction-worker`
- recognized but not implemented: `trainee`

### Pay

Canonical form:

```text
<id> pay <rate> <population-kind>
```

Examples:

```text
38 pay 0.125 unskilled-worker
38 pay 0.375 professional
38 pay 0.250 soldier
```

Parse-time rules:

- rate must be a non-negative decimal with up to three fractional digits
- allowed population kinds: `unemployable`, `unskilled-worker`, `professional`, `soldier`, `spy`, `construction-worker`

### Ration

Canonical form:

```text
<id> ration <percent>
```

Example:

```text
16 ration 50%
```

### Name

Ship/colony naming:

```text
<id> name <name>
```

Planet naming:

```text
<planet-ref> name <name>
```

Examples:

```text
39 name "Dragonfire"
5-12-38/2 name "Goldball Prime"
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
├── SetupOrder
├── BuildChangeOrder
├── MiningChangeOrder
├── TransferOrder
├── AssembleFactoryOrder
├── AssembleMineOrder
├── AssembleUnitsOrder
├── MoveOrbitOrder
├── MoveSystemOrder
├── DraftOrder
├── PayOrder
├── RationOrder
├── NameObjectOrder
└── NamePlanetOrder
```

### Order interface

```go
type OrderKind int

type Order interface {
    Kind()     OrderKind
    Phase()    int
    Validate() error
}
```

### Support types

```go
// UnitSpec is a parsed unit reference with an optional tech level.
type UnitSpec struct {
    Kind         UnitKind
    TechLevel    TechLevel
    HasTechLevel bool
}

// SetupTransfer is one item line inside a setup block.
type SetupTransfer struct {
    Quantity int
    Unit     UnitSpec
}

// BuildTarget names the output of a factory group or factory assembly.
// IsConsumerGoods is true when the text token was "consumer-goods";
// otherwise Unit carries the target unit spec.
type BuildTarget struct {
    IsConsumerGoods bool
    Unit            UnitSpec
}

// OrbitRef is a parsed orbit reference inside a system.
// StarSequence is empty for a simple orbit number (e.g., "6").
// For a star-qualified reference (e.g., "c-4"), StarSequence is "c".
type OrbitRef struct {
    StarSequence string // empty means primary star
    Orbit        int    // 1..10
}

// PlanetRef identifies a specific planet by system coordinates and orbit.
// Used for planet-naming orders.
type PlanetRef struct {
    Coords Coords
    Orbit  OrbitRef
}

// PayRate is a wage rate stored as fixed-point thousandths of a gold unit.
// 0.125 gold/turn is stored as 125.
type PayRate int
```

### Concrete order structs

```go
type SetupKind int

const (
    SetupShip   SetupKind = iota + 1
    SetupColony
)

// SetupOrder — Phase 4
type SetupOrder struct {
    SetupKind SetupKind
    SourceID  int           // existing ship or colony ID
    Transfers []SetupTransfer
}

// BuildChangeOrder — Phase 6
type BuildChangeOrder struct {
    SourceID int
    GroupID  int
    Target   BuildTarget
}

// MiningChangeOrder — Phase 7
type MiningChangeOrder struct {
    SourceID  int
    GroupID   int
    DepositID int
}

// TransferOrder — Phase 8
type TransferOrder struct {
    SourceID int
    Quantity int
    Unit     UnitSpec
    TargetID int
}

// AssembleFactoryOrder — Phase 9
// Assembles factory units into a factory group with a production target.
type AssembleFactoryOrder struct {
    SourceID int
    Quantity int
    Unit     UnitSpec
    Target   BuildTarget
}

// AssembleMineOrder — Phase 9
// Assembles mine units into a mining group assigned to a deposit.
type AssembleMineOrder struct {
    SourceID  int
    Quantity  int
    Unit      UnitSpec
    DepositID int
}

// AssembleUnitsOrder — Phase 9
// Assembles any other unit type (drives, life support, weapons, etc.).
type AssembleUnitsOrder struct {
    SourceID int
    Quantity int
    Unit     UnitSpec
}

// MoveOrbitOrder — Phase 14
// In-system jump to a different orbit around the same star.
type MoveOrbitOrder struct {
    ShipID      int
    Destination OrbitRef
}

// MoveSystemOrder — Phase 14
// Inter-system jump to the named system coordinates.
type MoveSystemOrder struct {
    ShipID      int
    Destination Coords
}

// DraftOrder — Phase 15
type DraftOrder struct {
    SourceID   int
    Quantity   int
    Population UnitKind // must be a draftable population kind
}

// PayOrder — Phase 16
type PayOrder struct {
    SourceID   int
    Rate       PayRate
    Population UnitKind // must be a payable population kind
}

// RationOrder — Phase 16
type RationOrder struct {
    SourceID int
    Percent  int // 0..100
}

// NameObjectOrder — Phase 19
// Names a ship or colony referenced by integer ID.
type NameObjectOrder struct {
    ObjectID int
    Name     string
}

// NamePlanetOrder — Phase 19
// Names a planet referenced by system coordinates and orbit.
type NamePlanetOrder struct {
    Planet PlanetRef
    Name   string
}
```

Design notes:

- `MoveOrbitOrder` and `MoveSystemOrder` are separate concrete types so the parser does not need a destination sum-type at every call site. Neither type commits to a particular ship-location storage shape — that is resolved in Sprint 16.
- `NameObjectOrder` and `NamePlanetOrder` are separate so object-ID naming (ships and colonies) and planet-location naming stay unambiguous.
- `SetupOrder` captures the full transfer list so execution can verify materials without re-parsing.
- All ID fields (`SourceID`, `TargetID`, `GroupID`, `DepositID`, `ShipID`, `ObjectID`) are plain `int` at parse time because the domain does not yet have typed IDs for every entity. Tasks 2 and later may replace them with the appropriate typed ID if it exists in `domain`.
- `PayRate` is fixed-point thousandths rather than `float64` to avoid floating-point comparison issues in validation and execution.
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

Execution-time validation belongs in the later turn engine and app/domain execution
paths, not in the parser.

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

The app-layer parse result should expose stable diagnostics with:

- `line`
- `code`
- `message`

Recommended stable diagnostic codes:

| Code | Meaning |
|---|---|
| `unknown_command` | the line does not begin with a recognized command form |
| `not_implemented` | the line matches a known historical command that v0 does not support |
| `syntax` | the command is recognized, but the field count or clause layout is wrong |
| `invalid_value` | a field is present but fails static validation |
| `unterminated_quote` | a quoted field is not closed before end of line |
| `unterminated_setup` | EOF reached before `end` closed a `setup` block |
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
| Set up | 4 |
| Build Change | 6 |
| Mining Change | 7 |
| Transfer | 8 |
| Assemble (all variants) | 9 |
| Move (both variants) | 14 |
| Draft | 15 |
| Pay | 16 |
| Ration | 16 |
| Name (both variants) | 19 |

Parse-time does not require players to submit orders in phase order. Input order is
preserved, but phase assignment is explicit on the typed order values.

## Compatibility Notes

- v0 canonical syntax follows the scrubbed, whitespace-based style already described in `apps/site/content/docs/developers/reference/agent-reference.md`.
- Historical comma-separated examples remain useful as source material, but they are not the grammar target for Sprint 12.
- The parser may accept a narrow alias set for convenience, but every new example added to the codebase should use the canonical forms in this document.
