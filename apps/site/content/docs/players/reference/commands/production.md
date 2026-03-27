---
title: Production
---

Production orders manage factories, mines, assembly, logistics, and the creation of new ships and colonies.

---

## Set up

Creates a new ship or colony by transferring materials from an existing ship or colony. Executes in phase 4 (Set up).

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
<id> build change <group-id> <build-target>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<group-id>` | Factory group number. Must be a positive integer. |
| `<build-target>` | Target unit type or `consumer-goods`. See [Units](/docs/players/reference/units/) for valid unit tokens. |

**Examples**

```text
16 build change 8 hyper-engine-1
16 build change 9 consumer-goods
```

---

## Mining Change

Reassigns a mining group to a different deposit. Executes in phase 7 (Mining change).

**Syntax**

```text
<id> mining change <group-id> <deposit-id>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<group-id>` | Mining group number. Must be a positive integer. |
| `<deposit-id>` | Target deposit ID. Must be a positive integer. |

**Example**

```text
348 mining change 18 92
```

---

## Transfer

Moves a quantity of one unit type from one ship or colony to another at the same location. Executes in phase 8 (Transfers).

One `transfer` order moves one unit type. Issue multiple orders to move several unit types in the same turn.

**Syntax**

```text
<id> transfer <quantity> <unit-token> <id>
```

**Parameters**

| Parameter | Description |
|---|---|
| First `<id>` | Source ship or colony ID. Must be a positive integer. |
| `<quantity>` | Number of units to transfer. Must be a positive integer. |
| `<unit-token>` | Unit type to transfer. See [Units](/docs/players/reference/units/) for valid tokens. |
| Second `<id>` | Destination ship or colony ID. Must be a positive integer. |

**Example**

```text
22 transfer 10 spy 29
```

---

## Assemble

Converts disassembled units into an assembled group. Executes in phase 9 (Assembly). The variant is determined by the unit token in the third field.

### Factory assembly

Assembles factory units into a factory group with a production target.

**Syntax**

```text
<id> assemble <quantity> <factory-unit> <build-target>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony or ship ID. Must be a positive integer. |
| `<quantity>` | Number of factory units to assemble. Must be a positive integer. |
| `<factory-unit>` | A factory unit token, e.g. `factory-6`. |
| `<build-target>` | Initial production target. A unit token or `consumer-goods`. |

**Example**

```text
91 assemble 54000 factory-6 consumer-goods
```

### Mine assembly

Assembles mine units into a mining group assigned to a specific deposit.

**Syntax**

```text
<id> assemble <quantity> <mine-unit> <deposit-id>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony or ship ID. Must be a positive integer. |
| `<quantity>` | Number of mine units to assemble. Must be a positive integer. |
| `<mine-unit>` | A mine unit token, e.g. `mine-2`. |
| `<deposit-id>` | Deposit to assign the group to. Must be a positive integer. |

**Example**

```text
83 assemble 25680 mine-2 148
```

### Other assembly

Assembles any other unit type (drives, life support, weapons, and similar).

**Syntax**

```text
<id> assemble <quantity> <unit-token>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony or ship ID. Must be a positive integer. |
| `<quantity>` | Number of units to assemble. Must be a positive integer. |
| `<unit-token>` | Unit type to assemble. See [Units](/docs/players/reference/units/) for valid tokens. |

**Example**

```text
58 assemble 6000 missile-launcher-1
```
