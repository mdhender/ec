---
title: Administration
---

Administration orders manage colony personnel, wages, food rationing, and the names of ships, colonies, and planets.

---

## Draft

Converts population into specialist roles. Executes in phase 15 (Draft).

**Syntax**

```text
<id> draft <quantity> <population-kind>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<quantity>` | Number of people to draft. Must be a positive integer. |
| `<population-kind>` | Target specialist kind. Accepted values: `professional`, `soldier`, `spy`, `construction-worker`. |

**Examples**

```text
13 draft 3600 soldier
16 draft 400 professional
16 draft 250 construction-worker
```

---

## Pay

Sets the wage rate for a population kind at a colony. Executes in phase 16 (Pay/ration).

**Syntax**

```text
<id> pay <rate> <population-kind>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<rate>` | Wage rate in gold per turn. A non-negative decimal with up to three fractional digits, e.g. `0.125`. |
| `<population-kind>` | Population kind to set wages for. Accepted values: `unemployable`, `unskilled-worker`, `professional`, `soldier`, `spy`, `construction-worker`. |

**Examples**

```text
38 pay 0.125 unskilled-worker
38 pay 0.375 professional
38 pay 0.250 soldier
```

---

## Ration

Sets the food ration percentage at a colony. Executes in phase 16 (Pay/ration).

**Syntax**

```text
<id> ration <percent>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Colony ID. Must be a positive integer. |
| `<percent>` | Ration level. An integer followed by `%`, in the range `0`–`100`. |

**Example**

```text
16 ration 50%
```

---

## Name

Assigns a name to a ship, colony, or planet. Executes in phase 19 (Naming/control). Names must be enclosed in double quotes and may be 1–24 characters.

### Name a ship or colony

**Syntax**

```text
<id> name "<name>"
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Ship or colony ID. Must be a positive integer. |
| `"<name>"` | New name. A quoted string, 1–24 characters. |

**Example**

```text
39 name "Dragonfire"
```

### Name a planet

**Syntax**

```text
<planet-ref> name "<name>"
```

**Parameters**

| Parameter | Description |
|---|---|
| `<planet-ref>` | Planet location as `<x>-<y>-<z>/<orbit-ref>`, e.g. `5-12-38/2` or `5-12-38/c-4`. |
| `"<name>"` | New name. A quoted string, 1–24 characters. |

**Example**

```text
5-12-38/2 name "Goldball Prime"
```
