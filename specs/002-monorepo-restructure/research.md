# Research: Monorepo Restructure

## 1. Go Module Path Preservation

**Decision**: Keep module path `arsen` in `apps/api/go.mod` and add a `go.work` file at the repo root with `use ./apps/api`.

**Rationale**: The module path `arsen` is a local/vanity name with no remote import expectations, so it works fine from any directory. A `go.work` file lets editors (VS Code, GoLand) and root-level tooling resolve the module automatically.

**Alternatives considered**:
- Changing module path to `arsen/apps/api` — rejected; unnecessary churn, breaks all internal imports
- Omitting `go.work` — rejected; editors and root-level Go tooling would not resolve the module

## 2. Next.js Scaffold

**Decision**: Target Next.js 16 (latest stable). Scaffold with `pnpm create next-app@latest apps/web --typescript --tailwind --eslint --app --src-dir --use-pnpm`.

**Rationale**: Next.js 16 is the current stable line, ships with Tailwind v4, Turbopack, and App Router by default. The `create-next-app` CLI with flags gives a reproducible, non-interactive setup.

**Alternatives considered**:
- Manual scaffolding — rejected; error-prone
- Next.js 15 — rejected; now in maintenance

## 3. Vitest + React Testing Library

**Decision**: Use Vitest with `@vitejs/plugin-react` and `happy-dom` environment. Install `@testing-library/react`, `@testing-library/jest-dom`, and `@testing-library/user-event`.

**Rationale**: Next.js official docs now recommend Vitest over Jest. `@vitejs/plugin-react` handles JSX/TSX transforms natively; `happy-dom` is faster than jsdom. Async Server Components require Playwright for E2E testing, not unit tests.

**Alternatives considered**:
- `next/jest` with Jest — rejected; slower cold starts, CJS-first, Jest is legacy path
- jsdom — rejected; slower than happy-dom with no real benefit for component tests

## 4. Docker Compose Multi-Service Build Contexts

**Decision**: Set each service's `build.context` to its workspace directory (`./apps/api`, `./apps/web`) with Dockerfiles in each workspace.

**Rationale**: Each Dockerfile only needs its own workspace files, keeping build contexts small and layer caching effective. If shared root files are needed later, change context to `.` and use `dockerfile: apps/api/Dockerfile`.

**Alternatives considered**:
- Single root context for all services — rejected; bloated build context
- Dockerfile at root referencing subdirectories — rejected; couples services unnecessarily

## 5. Root Taskfile Delegation

**Decision**: Use Taskfile v3 `includes:` with the `dir:` option:

```yaml
includes:
  api:
    taskfile: ./apps/api/Taskfile.yml
    dir: ./apps/api
  web:
    taskfile: ./apps/web/Taskfile.yml
    dir: ./apps/web
```

Invoke as `task api:build`, `task web:dev`, etc. Root-level aggregate tasks (e.g., `task test`) call both.

**Rationale**: The `dir:` key ensures each included Taskfile runs commands relative to its own directory. This is a first-class Taskfile v3 feature.

**Alternatives considered**:
- Shell `cd` wrappers in a single Taskfile — rejected; fragile, loses task isolation
- Turborepo — explicitly excluded by user choice

## 6. Monorepo .gitignore Strategy

**Decision**: One root `.gitignore` with sections for Go, Node/Next.js, and shared patterns. Keep workspace-level `.gitignore` files for tool-specific output (`.next/`, `bin/`).

**Rationale**: Git applies `.gitignore` rules hierarchically, so the root file covers shared patterns (`.env`, `.DS_Store`, IDE files) while workspace-level files handle local artifacts. Avoids duplication while maintaining workspace portability.

**Alternatives considered**:
- Only per-workspace ignores — rejected; duplicates common patterns
- Only root ignore — rejected; becomes unwieldy for tool-specific patterns
