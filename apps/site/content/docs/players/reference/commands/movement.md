---
title: Movement
---

Movement orders direct ships to change position. Both forms execute in phase 14 (Ship movement).

---

## Move (in-system)

Moves a ship to a different orbit within the same star system.

**Syntax**

```text
<id> move orbit <orbit-ref>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Ship ID. Must be a positive integer. |
| `<orbit-ref>` | Target orbit. A bare orbit number (`1`–`10`), or a star-qualified orbit (`<star-letter>-<orbit>`, e.g. `c-4`). |

**Examples**

```text
77 move orbit 6
88 move orbit c-4
```

---

## Move (system jump)

Moves a ship to a different star system.

**Syntax**

```text
<id> move system <system-coords>
```

**Parameters**

| Parameter | Description |
|---|---|
| `<id>` | Ship ID. Must be a positive integer. |
| `<system-coords>` | Destination system as `<x>-<y>-<z>`. Each component is an integer in `0`–`30`. |

**Example**

```text
79 move system 4-6-19
```
