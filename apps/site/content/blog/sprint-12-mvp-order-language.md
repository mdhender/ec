---
title: "Sprint 12: MVP Order Language"
date: 2026-03-27T15:00:00
---

{{< callout type="info" >}}
   Sprint 12 is complete. The API server can now parse MVP order text into typed domain values and return per-line diagnostics in a single pass.
{{< /callout >}}

## For Players

The full command reference is now published. Every order available in v0 — with syntax, parameters, and examples — is documented in the [Commands](/docs/players/reference/commands/) section of the player reference. If you've been waiting to know exactly how to write an assemble order or move a ship, that's the place to start.

---

## What We Built

Sprint 11 gave the frontend a real dashboard. Sprint 12 gives orders meaning — at least at parse time.

Before this sprint, order text was stored as-is and never inspected. Sprint 12 adds a new endpoint that reads a plain-text order submission, parses every line, and tells you what it understood and what it didn't:

```
POST /api/:empireNo/orders/parse
Authorization: Bearer <token>
Content-Type: text/plain
```

Response shape:

```json
{
  "ok": false,
  "accepted_count": 2,
  "diagnostics": [
    {
      "line": 3,
      "code": "not_implemented",
      "message": "bombard is not yet implemented"
    }
  ]
}
```

`ok` is `true` only when `diagnostics` is empty — meaning every non-blank, non-comment line produced a typed order. `accepted_count` is the number of lines that did. `diagnostics` is the list of lines that didn't, with a stable 1-based line number, a short code, and a human-readable message.

The endpoint returns `200` for any successful parse pass — even one with diagnostics. A partially-valid order file is not a server error; it's information. It returns `413` if the request body exceeds the same `maxOrderBytes` limit used by `POST /api/:empireNo/orders`, and `500` only for unexpected internal failures.

The input format is `text/plain`, one order per line. Blank lines and comment lines (starting with `;;`) are silently ignored.

This endpoint does **not** execute orders. It does not mutate game state. No turns are processed. Its job is to tell you whether your orders are syntactically and statically valid before you commit them.

---

## How It's Layered

The implementation follows the same SOUSA pattern used in every previous sprint — one concern per layer, each depending only inward.

**`domain/orders.go`** — the domain owns typed order values. Every MVP command has a concrete Go struct: `BuildChangeOrder`, `DraftOrder`, `PayOrder`, `RationOrder`, `TransferOrder`, `AssembleFactoryOrder`, `AssembleShipOrder`, `SetUpOrder`, `MoveOrder`, `NameOrder`, and so on. Each struct carries the parsed fields as typed values — `domain.UnitKind`, colony IDs, ship IDs, counts — rather than raw strings. An `OrderKind` enum identifies the command, and a `Phase` constant maps each order to its turn-processing phase. Pure validation methods (`Validate()`) check static invariants — positive IDs, non-empty names, percentages in range — without touching live game state.

**`app/order_parse_ports.go` and `app/order_parse_service.go`** — the app layer owns the `OrderParser` port, the `ParseOrdersService`, and the `ParseResult` / `ParseDiagnostic` types. The service accepts raw text, calls the port, and returns a stable result. It does not import `infra` or `delivery/http`. Delivery code can JSON-encode the result without knowing anything about the parser.

**`infra/ordertext/parser.go`** — the concrete line-oriented text parser lives here. It satisfies the `OrderParser` port using nothing but the Go standard library — no parser generator, no new dependency. It tokenizes each line, matches the first token against known command keywords, constructs the appropriate domain struct, and runs domain validation. Lines that fail produce a diagnostic with a descriptive code (`bad_syntax`, `invalid_value`, etc.). Lines for known but non-MVP commands — `bombard`, `invade`, `survey`, and others — produce a `not_implemented` diagnostic instead of a generic failure. Parsing always continues after a bad line so one request returns the full picture.

**`delivery/http/handlers.go`** — the `PostParseOrders` handler reads the request body with the same `maxOrderBytes` limit used by `PostOrders`, calls `ParseOrdersService`, and encodes the result as JSON. It maps an oversized body to `413` and an unexpected service error to `500`. HTTP shape only — no parsing logic here.

**`runtime/server/server.go`** — runtime constructs `ordertext.NewParser()`, wraps it in `app.NewParseOrdersService`, and passes the service into `delivery/http.AddRoutes`. This is the only layer that imports the concrete infra parser. Delivery never touches infra directly.

---

## What's Not Here Yet

- **Orders are parsed but not executed.** Accepted orders are returned in the response and then discarded. No turn pipeline yet.
- **No existence checks at parse time.** The parser does not verify that a referenced colony, ship, or deposit actually exists in the current game state. That validation belongs to the turn engine.
- **No ownership or reachability validation.** Whether the empire owns the target unit, or whether a move destination is reachable, is also deferred.
- **The frontend parse button hasn't been added.** Sprint 17 will wire the Orders page UI to call this endpoint. For now, `POST /orders/parse` is a backend-only contract.
- **Non-MVP commands return `not_implemented`.** Commands like `bombard`, `invade`, and `survey` are recognized but not parsed into typed orders.

---

## What's Next

The parse contract is now stable. Sprint 13 starts the turn engine foundation — the scaffolding that will eventually execute the orders this parser accepts. Sprint 14 builds the colony economy loop on top of that. By the time the frontend gets a parse button in Sprint 17, there will be a full pipeline waiting behind it.

---

## Version

The project is now at **v0.12.0-alpha**. The build is green, all backend tests pass, and `go vet` is clean.
