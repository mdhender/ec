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

3. **No mixed concerns.** Never combine semantic translation with cleanup or
   refactoring in the same task.

4. **Tasks must fit in context.** Each task description must be self-contained:
   an agent should be able to read the task and begin work without needing to
   read the entire repository. Include file paths, function names, expected
   behavior, and acceptance criteria inline.

5. **Leave the repo green.** Every completed task must leave all existing tests
   passing. If a task would break an earlier pass, it is scoped wrong.

6. **Small diffs only.** Prefer several small tasks over one large one. If a
   task will touch more than ~200 lines or more than 3 files, split it.

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

**Acceptance criteria:**
- [ ] [Observable behavior or test that must pass]
- [ ] [Second criterion]

**Tests to add/update:**
- `TestXxx` in `path/to/file_test.go` — [what it verifies]

---

### Task 2: [Short title]

**Subsystem:**
**Files:**
**Depends on:**

**What to do:**

**Acceptance criteria:**
- [ ]

**Tests to add/update:**
- `TestXxx` in `path/to/file_test.go` — [what it verifies]

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
