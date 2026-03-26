# Sprint 11: Frontend Dashboard Enhancement

**Pass:** Pass 3
**Goal:** Replace the minimal dashboard with data-driven cards (colonies, ships, planets), add sidebar links for colonies, ships, and star list, and implement the linked summary pages.
**Predecessor:** Sprint 10

---

## Sprint Rules

1. **One subsystem per task.** Each task targets exactly one bounded piece of
   work. If a task touches more than one subsystem, split it.

2. **Every task names its tests.** A task is not ready for an agent until it
   lists the exact tests to add or update. (Frontend: TypeScript compilation
   is the test — all tasks must leave `npm run build` passing.)

3. **No unrelated cleanup.** Do not bundle opportunistic refactoring into a
   feature task. Required follow-through cleanup belongs in the same task
   when it is directly caused by the change: stale references, dead helpers,
   guard clauses, invariant fixes, API alignment within the touched
   subsystem, and tests.

4. **Tasks must fit in context.** Each task description must be self-contained.
   Include file paths, component names, prop types, and expected behavior
   inline.

5. **Leave the repo green.** Every completed task must leave `npm run build`
   passing.

6. **Small diffs only.** Prefer several small tasks over one large one.

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

This sprint is **frontend only**. The backend API added in Sprint 10 is
already live at `GET /api/:empireNo/dashboard`. No backend changes are needed.

The app uses **no router**. Page state is a `Page` union type managed in
`App.tsx` via `useState` and `setPage()`. All navigation is done by calling
`setPage(...)`. New pages follow the same pattern.

Data fetching follows the `OrdersPage` pattern: `useEffect` + `useState`
for `loading`, `error`, and `data`. Do not add a third-party data-fetching
library.

After completing a task, update `sprints/sprint-11.md`: check off acceptance
criteria (change `[ ]` to `[x]`) and change the task status from TODO to DONE
in the Task Summary table at the bottom of the file.

### Current navigation (before this sprint)

```
Dashboard  (HomeIcon)
Orders     (ClipboardDocumentListIcon)
Reports    (DocumentTextIcon)
```

### Target navigation (after this sprint)

```
Dashboard  (HomeIcon)
Orders     (ClipboardDocumentListIcon)
Reports    (DocumentTextIcon)
Colonies   (BuildingOffice2Icon)
Ships      (RocketLaunchIcon)
Star List  (MapIcon)
```

All icons are from `@heroicons/react/24/outline`.

### Dashboard API response shape

```typescript
// from GET /api/:empireNo/dashboard
{
  colony_count:  number,
  colony_kinds:  Array<{ kind: string; count: number }>,
  ship_count:    number,
  planet_count:  number,
  planet_kinds:  Array<{ kind: string; count: number }>,
}
```

`colony_kinds` and `planet_kinds` omit entries with count 0.
`ship_count` is always 0 until ships are implemented.

### Page state after this sprint

```typescript
type Page =
  | "dashboard"
  | "orders"
  | "reports"
  | "report"
  | "admin-users"
  | "colonies"   // new
  | "ships"      // new
  | "star-list"; // new
```

### Key files

- `apps/web/src/lib/types.ts` — add `KindCount`, `DashboardSummary` (Task 1)
- `apps/web/src/lib/api.ts` — add `fetchDashboard` (Task 1)
- `apps/web/src/pages/ColoniesPage.tsx` — new file (Task 2)
- `apps/web/src/pages/ShipsPage.tsx` — new file (Task 2)
- `apps/web/src/pages/StarListPage.tsx` — new file (Task 2)
- `apps/web/src/App.tsx` — new pages, new nav items (Task 3)
- `apps/web/src/pages/DashboardPage.tsx` — cards, data fetch (Task 4)

### Build command

```bash
cd apps/web && npm run build
```

No automated test suite exists for the frontend. TypeScript compilation
success (`tsc -b`) is the acceptance gate for all tasks.

**Audit-left guidance:**
Audit items belong with the task that introduced the risk. SOUSA checks,
stale-reference scans, guard/negative tests, and API-pattern checks must
appear in the relevant task's design checklist or acceptance criteria. A
final audit task is only a last verification step — it should confirm
health, not discover new issues.

---

## Tasks

### Task 1: API types and client function

**Subsystem:** `lib/types.ts`, `lib/api.ts`
**Files:**
- `apps/web/src/lib/types.ts`
- `apps/web/src/lib/api.ts`
**Depends on:** None

**What to do:**

**1. Add to `lib/types.ts`:**

```typescript
export interface KindCount {
  kind: string;
  count: number;
}

export interface DashboardSummary {
  colony_count: number;
  colony_kinds: KindCount[];
  ship_count: number;
  planet_count: number;
  planet_kinds: KindCount[];
}
```

**2. Add to `lib/api.ts`:**

```typescript
export async function fetchDashboard(empireNo: number): Promise<DashboardSummary> {
  return apiFetch<DashboardSummary>(`/${empireNo}/dashboard`);
}
```

Add `DashboardSummary` to the import from `./types`.

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [ ] app
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: N/A — frontend only

_Existing pattern to follow:_
- `lib/types.ts` — existing interface definitions (e.g., `OrderEntry`)
- `lib/api.ts` — `apiFetch<T>` pattern used by `fetchOrders`, `fetchReports`

_Failure paths / guard clauses:_
- [ ] Not-found behavior specified (N/A — type definitions only)
- [ ] Empty/nil/invalid input behavior specified (N/A)
- [ ] ID/index bounds behavior specified (N/A)

_Invariants / validation:_
- [ ] Uniqueness / ID generation rule stated (N/A)
- [ ] Ordering or state preconditions stated (N/A)
- [ ] Validation rules listed and layer assigned (N/A — frontend types mirror backend)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: None — additive
- Search commands: N/A

**Acceptance criteria:**
- [ ] `cd apps/web && npm run build` succeeds
- [ ] `KindCount` and `DashboardSummary` are exported from `lib/types.ts`
- [ ] `fetchDashboard` is exported from `lib/api.ts`
- [ ] New/changed API matches an existing pattern (or deviation documented)

**Tests to add/update:**
- None beyond build success.

---

### Task 2: New summary pages

**Subsystem:** `pages/`
**Files:**
- `apps/web/src/pages/ColoniesPage.tsx` (new file)
- `apps/web/src/pages/ShipsPage.tsx` (new file)
- `apps/web/src/pages/StarListPage.tsx` (new file)
**Depends on:** Task 1

**What to do:**

Create three new page components. `ColoniesPage` fetches and displays real
summary data. `ShipsPage` and `StarListPage` are informational placeholders.

**1. `ColoniesPage.tsx`:**

Props: `{ empireNo: number }`

Fetches `fetchDashboard(empireNo)` on mount. Displays colony counts by kind.

```tsx
import { useEffect, useState } from "react";
import { fetchDashboard } from "../lib/api";
import type { DashboardSummary } from "../lib/types";

interface ColoniesPageProps {
  empireNo: number;
}

export default function ColoniesPage({ empireNo }: ColoniesPageProps) {
  const [data, setData] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchDashboard(empireNo)
      .then(setData)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [empireNo]);

  if (loading) return <p className="text-gray-500">Loading…</p>;
  if (error) return <p className="text-red-600">{error}</p>;
  if (!data || data.colony_count === 0) {
    return (
      <div>
        <h1 className="text-2xl font-semibold text-gray-900 mb-4">Colonies</h1>
        <p className="text-gray-500">No colonies.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Colonies</h1>
      <p className="text-sm text-gray-500 mb-4">
        {data.colony_count} {data.colony_count === 1 ? "colony" : "colonies"} total
      </p>
      <table className="min-w-full divide-y divide-gray-200">
        <thead>
          <tr>
            <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">Kind</th>
            <th className="px-4 py-2 text-right text-sm font-medium text-gray-500">Count</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {data.colony_kinds.map((kc) => (
            <tr key={kc.kind}>
              <td className="px-4 py-2 text-sm text-gray-900">{kc.kind}</td>
              <td className="px-4 py-2 text-sm text-gray-900 text-right">{kc.count}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

**2. `ShipsPage.tsx`:**

Props: none

Static placeholder page.

```tsx
export default function ShipsPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Ships</h1>
      <p className="text-gray-500">
        No ships. (The assemble ship order has not been implemented.)
      </p>
    </div>
  );
}
```

**3. `StarListPage.tsx`:**

Props: none

Static placeholder page. This page shows stars the empire has probed.
Probing is not yet implemented.

```tsx
export default function StarListPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900 mb-4">Star List</h1>
      <p className="text-gray-500">
        No stars. (The probe order has not been implemented.)
      </p>
    </div>
  );
}
```

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [ ] app
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: N/A — frontend only

_Existing pattern to follow:_
- `OrdersPage.tsx` — `useEffect` + `useState` data-fetching pattern
- `ReportsPage.tsx` — loading/error state handling

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: `ColoniesPage` shows "No colonies." when `colony_count == 0`
- [x] Empty/nil/invalid input behavior specified: loading and error states handled
- [ ] ID/index bounds behavior specified (N/A)

_Invariants / validation:_
- [ ] Uniqueness / ID generation rule stated (N/A)
- [ ] Ordering or state preconditions stated (N/A)
- [ ] Validation rules listed and layer assigned (N/A — frontend display only)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: None — new files
- Search commands: N/A

**Acceptance criteria:**
- [ ] `cd apps/web && npm run build` succeeds
- [ ] `ColoniesPage` fetches dashboard data and renders colony counts by kind
- [ ] `ColoniesPage` shows "No colonies." when `colony_count == 0`
- [ ] `ColoniesPage` handles loading and error states
- [ ] `ShipsPage` renders the correct placeholder message
- [ ] `StarListPage` renders the correct placeholder message
- [ ] New/changed API matches an existing pattern (or deviation documented)

**Tests to add/update:**
- None beyond build success.

---

### Task 3: App routing and sidebar navigation

**Subsystem:** `App.tsx`
**Files:**
- `apps/web/src/App.tsx`
**Depends on:** Task 2

**What to do:**

Add three new pages to the `Page` type, import and render the new page
components, and add three new nav items to the sidebar.

**1. Extend the `Page` type:**

```typescript
type Page =
  | "dashboard"
  | "orders"
  | "reports"
  | "report"
  | "admin-users"
  | "colonies"
  | "ships"
  | "star-list";
```

**2. Import new icons and pages:**

```typescript
import {
  HomeIcon,
  DocumentTextIcon,
  ClipboardDocumentListIcon,
  BuildingOffice2Icon,  // new
  RocketLaunchIcon,     // new
  MapIcon,              // new
} from "@heroicons/react/24/outline";

import ColoniesPage from "./pages/ColoniesPage";
import ShipsPage from "./pages/ShipsPage";
import StarListPage from "./pages/StarListPage";
```

**3. Add new nav items** to the `navigation` array after "Reports":

```typescript
{
  name: "Colonies",
  href: "#",
  icon: BuildingOffice2Icon,
  current: page === "colonies",
  onClick: () => setPage("colonies"),
},
{
  name: "Ships",
  href: "#",
  icon: RocketLaunchIcon,
  current: page === "ships",
  onClick: () => setPage("ships"),
},
{
  name: "Star List",
  href: "#",
  icon: MapIcon,
  current: page === "star-list",
  onClick: () => setPage("star-list"),
},
```

**4. Add new cases to `renderPage()`:**

```typescript
case "colonies":
  return <ColoniesPage empireNo={empireNo} />;
case "ships":
  return <ShipsPage />;
case "star-list":
  return <StarListPage />;
```

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [ ] app
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: N/A — frontend only

_Existing pattern to follow:_
- Existing `navigation` array items in `App.tsx` (Dashboard, Orders, Reports)
- `renderPage()` switch pattern in `App.tsx`

_Failure paths / guard clauses:_
- [ ] Not-found behavior specified (N/A — static page mapping)
- [ ] Empty/nil/invalid input behavior specified (N/A)
- [ ] ID/index bounds behavior specified (N/A)

_Invariants / validation:_
- [ ] Uniqueness / ID generation rule stated (N/A)
- [x] Ordering or state preconditions stated: nav items added after "Reports"
- [ ] Validation rules listed and layer assigned (N/A)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `Page` type, `renderPage()`, `navigation` array
- Search commands: `grep -n 'Page' apps/web/src/App.tsx`

**Acceptance criteria:**
- [ ] `cd apps/web && npm run build` succeeds
- [ ] `Page` type includes `"colonies"`, `"ships"`, `"star-list"`
- [ ] Sidebar has Colonies, Ships, and Star List nav items with correct icons
- [ ] Each new nav item highlights when its page is active
- [ ] `renderPage()` renders the correct component for each new page
- [ ] Stale references/helpers caused by this change removed or explicitly retained with reason
- [ ] New/changed API matches an existing pattern (or deviation documented)

**Tests to add/update:**
- None beyond build success.

---

### Task 4: Dashboard cards

**Subsystem:** `DashboardPage.tsx`
**Files:**
- `apps/web/src/pages/DashboardPage.tsx`
- `apps/web/src/App.tsx` (update props passed to DashboardPage)
**Depends on:** Tasks 1, 3

**What to do:**

Replace the current button-only dashboard with a card grid that shows
live colony, ship, and planet summary data fetched from the dashboard API.
Keep the existing Orders and Reports buttons below the cards.

**1. Rewrite `DashboardPage.tsx`:**

New props:

```typescript
interface DashboardPageProps {
  empireName: string;
  empireNo: number;
  onNavigateOrders: () => void;
  onNavigateReports: () => void;
  onNavigateColonies: () => void;
  onNavigateShips: () => void;
}
```

The component fetches `fetchDashboard(empireNo)` on mount using the
`OrdersPage` loading pattern (`useEffect`, `useState` for `loading`,
`error`, `data`).

**Card layout:** a responsive three-column grid:
```tsx
<div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
  {/* Colonies card */}
  {/* Ships card */}
  {/* Planets card */}
</div>
```

Each card:
```tsx
<div className="bg-white rounded-lg shadow p-6">
  <h3 className="text-sm font-medium text-gray-500 uppercase tracking-wide">
    {title}
  </h3>
  <p className="text-3xl font-bold text-gray-900 mt-2">{count}</p>
  <ul className="mt-2 space-y-1">
    {kinds.map((kc) => (
      <li key={kc.kind} className="text-sm text-gray-600">
        {kc.count} {kc.kind}
      </li>
    ))}
  </ul>
  {onNavigate && (
    <button
      onClick={onNavigate}
      className="mt-4 text-sm text-indigo-600 hover:text-indigo-800 font-medium"
    >
      View details →
    </button>
  )}
</div>
```

**Colonies card:** title "Colonies", count `data.colony_count`,
kinds `data.colony_kinds`, `onNavigate={onNavigateColonies}`.

**Ships card:** title "Ships", count `data.ship_count` (always 0),
kinds `[]` (no breakdown needed), `onNavigate={onNavigateShips}`.

**Planets card:** title "Planets", count `data.planet_count`,
kinds `data.planet_kinds`, no `onNavigate` (no planets page yet).

While loading, render a skeleton placeholder instead of the grid:
```tsx
<div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
  {[0, 1, 2].map((i) => (
    <div key={i} className="bg-white rounded-lg shadow p-6 animate-pulse">
      <div className="h-4 bg-gray-200 rounded w-1/2 mb-3" />
      <div className="h-8 bg-gray-200 rounded w-1/4" />
    </div>
  ))}
</div>
```

On error, render a brief inline error message above the cards area:
```tsx
<p className="text-sm text-red-600 mb-4">{error}</p>
```

The existing Orders and Reports buttons remain below the card grid,
unchanged.

**2. Update `App.tsx`** — pass new props to `DashboardPage`:

```tsx
<DashboardPage
  empireName={empireName}
  empireNo={empireNo}
  onNavigateOrders={() => setPage("orders")}
  onNavigateReports={() => setPage("reports")}
  onNavigateColonies={() => setPage("colonies")}
  onNavigateShips={() => setPage("ships")}
/>
```

**Design review checklist:**

_SOUSA layers touched:_
- [ ] domain
- [ ] app
- [ ] infra
- [ ] delivery
- [ ] runtime
- Allowed dependency direction: N/A — frontend only

_Existing pattern to follow:_
- `OrdersPage.tsx` — `useEffect` + `useState` data-fetching pattern
- Existing `DashboardPage.tsx` props pattern (extends with new callbacks)

_Failure paths / guard clauses:_
- [x] Not-found behavior specified: error message rendered inline on fetch failure
- [x] Empty/nil/invalid input behavior specified: skeleton placeholders during loading
- [ ] ID/index bounds behavior specified (N/A)

_Invariants / validation:_
- [ ] Uniqueness / ID generation rule stated (N/A)
- [x] Ordering or state preconditions stated: existing Orders/Reports buttons preserved below cards
- [ ] Validation rules listed and layer assigned (N/A — frontend display only)

_Impact scan:_
- Helpers/call sites/fields/comments/tests to revisit: `DashboardPage` props in `App.tsx` must be updated
- Search commands: `grep -n 'DashboardPage' apps/web/src/App.tsx`

**Acceptance criteria:**
- [ ] `cd apps/web && npm run build` succeeds
- [ ] `DashboardPage` accepts `empireNo` and the four `onNavigate*` callbacks
- [ ] Dashboard fetches `GET /api/:empireNo/dashboard` on mount
- [ ] Three cards render with correct title, count, and kind breakdown
- [ ] Colonies and Ships cards have "View details →" links; Planets card does not
- [ ] Skeleton placeholders render while loading
- [ ] Error message renders on fetch failure
- [ ] Existing Orders and Reports buttons are preserved below the cards
- [ ] `App.tsx` passes all required props to `DashboardPage`
- [ ] Stale references/helpers caused by this change removed or explicitly retained with reason
- [ ] New/changed API matches an existing pattern (or deviation documented)

**Tests to add/update:**
- None beyond build success.

---

### Task 5: Audit and build

**Subsystem:** all
**Files:** all files touched in Tasks 1–4
**Depends on:** Tasks 1–4

**What to do:**

1. **Unused imports** — check all modified files for imports that are no
   longer referenced after the changes. TypeScript's `tsc` will flag these
   if `"noUnusedLocals": true` is set; fix any that appear.

2. **Prop consistency** — verify the props passed to `DashboardPage` in
   `App.tsx` exactly match the `DashboardPageProps` interface in
   `DashboardPage.tsx`.

3. **Placeholder text** — verify the exact placeholder strings in
   `ShipsPage` and `StarListPage` match this document:
   - Ships: `"No ships. (The assemble ship order has not been implemented.)"`
   - Star List: `"No stars. (The probe order has not been implemented.)"`

4. **Full build:**
   ```bash
   cd apps/web && npm run build
   ```
   Fix any TypeScript errors or Vite build failures.

**Acceptance criteria:**
- [ ] No unused imports in modified files
- [ ] `DashboardPage` props in `App.tsx` match the component interface exactly
- [ ] Placeholder strings match the spec
- [ ] `cd apps/web && npm run build` succeeds with zero errors

**Tests to add/update:**
- None.

---

## Task Summary

| Task | Title                              | Status | Depends On | Agent/Thread | Notes |
|------|------------------------------------|--------|------------|--------------|-------|
| 1    | API types and client function      | TODO   | —          |              |       |
| 2    | New summary pages                  | TODO   | 1          |              |       |
| 3    | App routing and sidebar navigation | TODO   | 2          |              |       |
| 4    | Dashboard cards                    | TODO   | 1, 3       |              |       |
| 5    | Audit and build                    | TODO   | 1–4        |              |       |
