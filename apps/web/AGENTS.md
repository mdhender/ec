# Frontend — React SPA

Stack: React, Vite, Tailwind CSS, TypeScript (strict mode).

This is a static SPA. It is built to serve static files and served by the Go backend. Node.js is not available on the production server — there is no SSR, no server-side Node process.

Full architecture rules are in `docs/SOUSA.md` at the repo root.

## Directory Layout

```
src/
  lib/              API client, auth context, shared types, data-access helpers.
  components/       Reusable presentational UI components.
  pages/            Route/page modules — composition and page-level data flow.
```

## Key Rules

- **`src/lib/`** owns all API fetch logic, auth/token handling, shared DTOs, and common utilities. Do not scatter fetch or auth code across pages and components.
- **`src/components/`** should be mostly presentational. No direct backend calls, business rules, or auth policy in leaf components.
- **`src/pages/`** are delivery code. They compose lib and components. Do not let pages become a dumping ground for duplicated logic — extract to `src/lib/` instead.
- The backend is the single source of truth for business rules and authorization. Frontend may reflect permissions in the UI, but never rely on hidden buttons or route guards as the only enforcement.
- Avoid duplicating domain rules in TypeScript unless needed for UX, and keep it minimal.

## Dev Server

Vite proxies `/api` requests to `http://localhost:?` (the Go backend) during development. See `vite.config.ts`.

## Build

`npm run build` produces static assets in `dist/`. These are served by the Caddy server/Go backend in production.
