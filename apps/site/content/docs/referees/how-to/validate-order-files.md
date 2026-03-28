---
title: Validate Player Order Files
---

Use `cli parse orders` to check one or more order files for syntax errors before running a turn.
The command parses each file, reports accepted order counts and any diagnostics, and exits with a non-zero status if problems are found.

## Prerequisites

- The `cli` binary is built and on your path.
- You have one or more order files submitted by players (plain text, UTF-8).

## Validate a single file

```sh
cli parse orders orders-empire1.txt
```

Clean output looks like:

```text
orders-empire1.txt: 14 orders, 0 diagnostics
```

## Validate multiple files at once

Pass all files as arguments:

```sh
cli parse orders orders-empire1.txt orders-empire2.txt orders-empire3.txt
```

Each file is reported separately:

```text
orders-empire1.txt: 14 orders, 0 diagnostics
orders-empire2.txt: 9 orders, 0 diagnostics
orders-empire3.txt: 12 orders, 0 diagnostics
```

## Read the diagnostics

When a file contains errors, the command prints each diagnostic indented below the file summary with the line number, diagnostic code, and a human-readable message:

```text
orders-empire4.txt: 3 orders, 4 diagnostics
  line 5 [unknown_command]: unrecognized command "frobnicator"
  line 7 [syntax]: build change requires <colonyID> <groupNo> <unitKind>
  line 10 [unterminated_quote]: unterminated quoted string
  line 12 [invalid_value]: assemble: invalid unit kind "hyper-engine": "hyper-engine" requires a tech-level suffix (e.g. hyper-engine-1)
```

Valid orders in the same file are still counted — the parser does not stop at the first error.

## Check the exit status

The command exits with status 1 if any file had read errors or diagnostics.
Use this in a script to gate turn processing:

```sh
if cli parse orders orders-*.txt; then
    echo "all orders clean"
else
    echo "fix errors before running the turn"
fi
```

## Diagnostic codes

| Code | Meaning |
|---|---|
| `unknown_command` | Line does not begin with a recognized command |
| `not_implemented` | Recognized command that v0 does not yet support |
| `syntax` | Recognized command with wrong field count or layout |
| `invalid_value` | A field is present but fails static validation |
| `unterminated_quote` | A quoted string is not closed before end of line |
| `unexpected_end` | `end` appeared outside an open `setup` block |
