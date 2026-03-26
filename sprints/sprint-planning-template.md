# Sprint N: [Title]

**Pass:** [Pass N]
**Goal:** [One sentence describing the sprint's deliverable]
**Predecessor:** [Link to prior sprint or "None"]

---

## Sprint Rules

1. **One subsystem per task.** Each task targets exactly one bounded piece of
   work. If a task touches more than one subsystem, split it.

2. **Every task names its tests.** A task is not ready for an agent until it
   lists the exact tests to add or update.

3. **No unrelated cleanup.** Do not bundle opportunistic refactoring into a
   feature task. Required follow-through cleanup belongs in the same task
   when it is directly caused by the change: stale references, dead helpers,
   guard clauses, invariant fixes, API alignment within the touched
   subsystem, and tests.

4. **Tasks must fit in context.** Each task description must be self-contained:
   an agent should be able to read the task and begin work without needing to
   read the entire repository. Include file paths, function names, expected
   behavior, and acceptance criteria inline.

5. **Leave the repo green.** Every completed task must leave all existing tests
   passing. If a task would break an earlier pass, it is scoped wrong.

6. **Small diffs only.** Prefer several small tasks over one large one. If a
   task will touch more than ~200 lines or more than 3 files, split it.

7. **Every task must state failure paths and invariants.** If a task does
   lookup, selection, indexing, parsing, ID allocation, or file/template
   input, the task must define behavior for not-found / invalid / empty /
   duplicate cases and name tests for them.

8. **Every task must include an impact scan.** List the existing helpers,
   fields, comments, call sites, and tests that may become stale because of
   the change. Remove/update them in the same task, or explicitly say why
   they remain.

9. **New/changed APIs must match an existing pattern.** For ports, stores,
   constructors, and method signatures, the task must cite the existing
   pattern it follows. If it deviates, the task must briefly justify the
   deviation.

10. **Validation ownership must be explicit.** For external inputs
    (JSON/templates/CLI/API payloads), the task must say which SOUSA layer
    validates invariants (`domain` vs `app`) and what is validated.

---

## Context for Agents

<!-- Paste or summarize the minimum context an agent needs to work on ANY task
     in this sprint. Keep this short — it is included so agents do not have to
     discover it themselves. -->

After completing a task, update sprints/sprint-[ThisSprint].md: check off acceptance criteria (change [ ] to [x]) and change the task status from TODO to DONE in the Task Summary table at the bottom of the file.

**Key files:**
- (list the files agents will read or edit)

**Key types/functions:**
- (list the specific types, functions, or subsystems in scope)

**Build/test commands:**
```bash
# Build C reference
cmake -S . -B build && cmake --build build

# Run Go tests
go test ./...

# Run differential suite (if applicable)
go test ./internal/diff/...
```

**Constraints reminder:**
- stdlib first — minimal third-party dependencies
- tests default to in-memory databases when possible

**Audit-left guidance:**
Audit items belong with the task that introduced the risk. SOUSA checks,
stale-reference scans, guard/negative tests, and API-pattern checks must
appear in the relevant task's design checklist or acceptance criteria. A
final audit task is only a last verification step — it should confirm
health, not discover new issues.

---

## Tasks

### Task 1: [Short title]

**Subsystem:** [e.g., "CLI option parsing" or "symbol table"]
**Files:** [exact paths to read and modify]
**Depends on:** None | Task N

**What to do:**
<!-- 3–10 lines. Be specific: name functions, describe inputs/outputs,
     state the expected behavior. Do not say "implement X" without explaining
     what X does. -->

**Design review checklist:**
<!-- Fill in during planning. Use N/A for items that don't apply. -->

_SOUSA layers touched:_
- [ ] domain
- [ ] app
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: [e.g., "app → domain only"]

_Existing pattern to follow:_
- [Comparable file/function/interface to mirror]
- [If different, why]

_Failure paths / guard clauses:_
- [ ] Not-found behavior specified
- [ ] Empty/nil/invalid input behavior specified
- [ ] ID/index bounds behavior specified (if applicable)

_Invariants / validation:_
- [ ] Uniqueness / ID generation rule stated
- [ ] Ordering or state preconditions stated
- [ ] Validation rules listed and layer assigned (`domain` or `app`)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: [list or "None"]
- Search commands: [e.g., `rg 'FuncName' backend/internal/`]

**Acceptance criteria:**
- [ ] [Observable behavior or test that must pass]
- [ ] [Second criterion]
- [ ] At least one negative/guard-path test added or updated (if task has lookups, parsing, or ID allocation)
- [ ] Stale references/helpers caused by this change removed or explicitly retained with reason
- [ ] New/changed API matches an existing pattern (or deviation documented)
- [ ] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestXxx_HappyPath` in `path/to/file_test.go` — [what it verifies]
- `TestXxx_InvalidOrNotFound` in `path/to/file_test.go` — [guard/failure case]
- Existing tests updated: [list or "None"]

---

### Task 2: [Short title]

**Subsystem:**
**Files:**
**Depends on:**

**What to do:**

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [ ] app
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction:

_Existing pattern to follow:_
-

_Failure paths / guard clauses:_
- [ ] Not-found behavior specified
- [ ] Empty/nil/invalid input behavior specified
- [ ] ID/index bounds behavior specified (if applicable)

_Invariants / validation:_
- [ ] Uniqueness / ID generation rule stated
- [ ] Ordering or state preconditions stated
- [ ] Validation rules listed and layer assigned (`domain` or `app`)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit:
- Search commands:

**Acceptance criteria:**
- [ ]
- [ ] At least one negative/guard-path test added or updated (if applicable)
- [ ] Stale references/helpers caused by this change removed or explicitly retained
- [ ] New/changed API matches an existing pattern (or deviation documented)
- [ ] SOUSA boundary valid for touched layers

**Tests to add/update:**
- `TestXxx_HappyPath` in `path/to/file_test.go` — [what it verifies]
- `TestXxx_InvalidOrNotFound` in `path/to/file_test.go` — [guard/failure case]
- Existing tests updated: [list or "None"]

---

<!-- Copy the Task block above for additional tasks. -->

---

## Task Summary

<!-- Agents update their task status here when work is complete.
     Valid statuses: TODO → IN-PROGRESS → DONE | BLOCKED -->

| Task | Title                | Status      | Agent/Thread | Notes |
|------|----------------------|-------------|--------------|-------|
| 1    | [Short title]        | TODO        |              |       |
| 2    | [Short title]        | TODO        |              |       |
