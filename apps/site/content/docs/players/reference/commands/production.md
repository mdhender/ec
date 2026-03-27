---
title: Production
---

Production orders manage factories, mines, assembly, logistics, and the creation of new ships and colonies.

---

## Set up

Creates a new ship or colony by transferring materials from an existing ship or colony. Executes in phase 4 (Set up).

{{< callout type="warning" >}}
   Set up is recognized by the parser but returns `not_implemented` in v0. The full multi-line block syntax will be enabled in a later release.
{{< /callout >}}

`setup` is the only multi-line order. The block opens with a `setup` line, contains one or more `transfer` lines, and closes with `end`.

**Syntax**

```text
setup ship from <id>
transfer <quantity> <unit-token>
...
end
```

```text
setup colony from <id>
transfer <quantity> <unit-token>
...
end
```

**Parameters**

| Parameter | Description |
|---|---|
| `ship` / `colony` | The kind of entity to create. |
| `<id>` | Source ship or colony ID. Must be a positive integer. |
| `<quantity>` | Number of units to transfer. Must be a positive integer. |
| `<unit-token>` | Unit type to transfer. See [Units](/docs/players/reference/units/) for valid tokens. |

The new ship or colony receives an ID assigned during execution; it is not specified in the order.

**Example**

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

---

## Build Change

Redirects a factory group to produce a different unit type. Executes in phase 6 (Build change).

**Syntax**

```text
build change <id> <group-id> <build-target>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<group-id>` | Factory group number. Must be a positive integer. |
| `<build-target>` | Target unit type or `consumer-goods`. See [Units](/docs/players/reference/units/) for valid unit tokens. |

**Examples**

```text
build change 16 8 hyper-engine-1
build change 16 9 consumer-goods
```

---

## Mining Change

Reassigns a mining group to a different deposit. Executes in phase 7 (Mining change).

**Syntax**

```text
mining change <id> <group-id> <deposit-id>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<group-id>` | Mining group number. Must be a positive integer. |
| `<deposit-id>` | Target deposit ID. Must be a positive integer. |

**Example**

```text
mining change 348 18 92
```

---

## Transfer

Moves a quantity of one unit type from one ship or colony to another at the same location. Executes in phase 8 (Transfers).

One `transfer` order moves one unit type. Issue multiple orders to move several unit types in the same turn.

**Syntax**

```text
transfer <source-id> <dest-id> <unit-token> <quantity>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<source-id>` | Source ship or colony ID. Must be a positive integer. |
| `<dest-id>` | Destination ship or colony ID. Must be a positive integer. |
| `<unit-token>` | Unit type to transfer. See [Units](/docs/players/reference/units/) for valid tokens. |
| `<quantity>` | Number of units to transfer. Must be a positive integer. |

**Example**

```text
transfer 22 29 spy 10
```

---

## Assemble

Converts disassembled units into an assembled group. Executes in phase 9 (Assembly).

There are three forms. The token after the location ID determines which form is used.

### Assemble generic units

Assembles any unit type other than factories or mines.

**Syntax**

```text
assemble <id> <unit-token> <quantity>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony or ship ID. Must be a positive integer. |
| `<unit-token>` | Unit type to assemble. See [Units](/docs/players/reference/units/) for valid tokens. |
| `<quantity>` | Number of units to assemble. Must be a positive integer. Commas are accepted. |

**Examples**

```text
assemble 58 missile-launcher-1 6000
assemble 58 missile-launcher-1 6,000
```

### Assemble factories

Assembles factory units and sets what they will produce.

**Syntax**

```text
assemble <id> factory <factory-unit> <quantity> <build-target>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony or ship ID. Must be a positive integer. |
| `<factory-unit>` | Factory unit type, e.g. `factory-6`. Must include a tech-level suffix. |
| `<quantity>` | Number of factory units to assemble. Must be a positive integer. Commas are accepted. |
| `<build-target>` | Unit type the factories will produce. See [Units](/docs/players/reference/units/) for valid tokens. |

**Examples**

```text
assemble 91 factory factory-6 54000 hyper-engine-1
assemble 91 factory factory-6 54,000 hyper-engine-1
assemble 16 factory factory-4 12000 consumer-goods
```

### Assemble mines

Assembles mine units and assigns them to a deposit.

**Syntax**

```text
assemble <id> mine <mine-unit> <quantity> <deposit-id>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony or ship ID. Must be a positive integer. |
| `<mine-unit>` | Mine unit type, e.g. `mine-2`. Must include a tech-level suffix. |
| `<quantity>` | Number of mine units to assemble. Must be a positive integer. Commas are accepted. |
| `<deposit-id>` | ID of the deposit to assign the mines to. Must be a positive integer. |

**Examples**

```text
assemble 83 mine mine-2 25680 92
assemble 83 mine mine-2 25,680 92
assemble 47 mine mine-1 4000 7
```
