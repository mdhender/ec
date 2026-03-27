---
title: Administration
---

Administration orders manage colony personnel, wages, food rationing, and the names of ships, colonies, and planets.

---

## Draft

Converts population into specialist roles. Executes in phase 15 (Draft).

**Syntax**

```text
draft <id> <population-kind> <quantity>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<population-kind>` | Target specialist kind. Accepted values: `professional`, `soldier`, `spy`, `construction-worker`. |
| `<quantity>` | Number of people to draft. Must be a positive integer. |

**Examples**

```text
draft 13 soldier 3600
draft 16 professional 400
draft 16 construction-worker 250
```

---

## Pay

Sets the wage rate for a population kind at a colony. Executes in phase 16 (Pay/ration).

**Syntax**

```text
pay <id> <population-kind> <rate>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<population-kind>` | Population kind to set wages for. Accepted values: `unemployable`, `unskilled-worker`, `professional`, `soldier`, `spy`, `construction-worker`. |
| `<rate>` | Wage rate in gold per turn. A non-negative integer. |

**Examples**

```text
pay 38 unskilled-worker 1
pay 38 professional 5
pay 38 soldier 4
```

---

## Ration

Sets the food ration percentage at a colony. Executes in phase 16 (Pay/ration).

**Syntax**

```text
ration <id> <percent>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<percent>` | Ration level as an integer in the range `0`–`100`. |

**Example**

```text
ration 16 50
```

---

## Name

Assigns a name to a ship, colony, or planet. Executes in phase 19 (Naming/control). The target kind (`ship`, `colony`, or `planet`) must be specified explicitly. Names must be enclosed in double quotes.

### Name a ship

**Syntax**

```text
name ship <id> <name>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Ship ID. Must be a positive integer. |
| `<name>` | New name. A quoted string of 1–24 characters. |

**Example**

```text
name ship 39 "Dragonfire"
name ship 39 "Slash // Burn"
```

### Name a colony

**Syntax**

```text
name colony <id> <name>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<name>` | New name. A quoted string of 1–24 characters. |

**Example**

```text
name colony 7 "Outpost Beta"
```

### Name a planet

**Syntax**

```text
name planet <id> <name>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Planet ID. Must be a positive integer. |
| `<name>` | New name. A quoted string of 1–24 characters. |

**Example**

```text
name planet 5 "New Terra"
```
