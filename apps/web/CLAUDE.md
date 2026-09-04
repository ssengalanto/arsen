@AGENTS.md

# Web App (apps/web/)

Next.js 16 frontend with React 19, TypeScript, Tailwind CSS v4, and App Router.

## Next.js 16 Warning

This project uses Next.js 16 which has breaking changes from earlier versions. Before writing any Next.js code, read the relevant guide in `node_modules/next/dist/docs/` — your training data may not reflect the current API surface.

## Commands

All commands run from this directory or via `task web:<cmd>` from the repo root.

```bash
pnpm dev          # Dev server (or: task web:dev)
pnpm build        # Production build (or: task web:build)
pnpm test         # Vitest (or: task web:test)
pnpm test:watch   # Vitest watch mode (or: task web:test-watch)
pnpm lint         # ESLint (or: task web:lint)
pnpm typecheck    # TypeScript strict check (or: task web:typecheck)
pnpm format       # Prettier (or: task web:format)
```

## Project Structure

```
src/
  app/
    layout.tsx      # Root layout
    page.tsx        # Home page
    globals.css     # Global styles (Tailwind v4)
__tests__/          # Test files (Vitest + React Testing Library + happy-dom)
```

Path alias: `@/*` maps to `./src/*`.

## Stack

- **Next.js 16** — App Router, standalone output mode
- **React 19** — Latest React with server components
- **TypeScript** — Strict mode enabled
- **Tailwind CSS v4** — Via `@tailwindcss/postcss`
- **Vitest 5** — Test runner with `happy-dom` environment
- **React Testing Library** — Component testing
- **ESLint** — `eslint-config-next` (core-web-vitals + typescript)
- **Prettier** — Code formatting
- **pnpm 10** — Package manager

## Testing

Tests use Vitest with React Testing Library and happy-dom. Test files live in `__tests__/`.

## Notes
- The standalone output mode is configured for Docker deployment.