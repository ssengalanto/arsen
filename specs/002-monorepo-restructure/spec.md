# Feature Specification: Monorepo Restructure

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "right now what we have is a single monolith but what we need it a monorepo 1 for api and 1 for web - which is going to be react typescript next.js"

## Clarifications

### Session 2026-09-04

- Q: What directory layout should the monorepo use for the two workspaces? → A: Both workspaces in `apps/api/` and `apps/web/` (conventional monorepo layout).
- Q: Should the monorepo use Turborepo for cross-workspace build orchestration, or rely solely on the existing Taskfile? → A: Taskfile-only at root orchestrates both workspaces; no Turborepo.
- Q: Which testing framework should the web workspace use? → A: Vitest + React Testing Library.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Monorepo Project Structure (Priority: P1)

As a developer, I can work on the API and web frontend as separate workspaces within a single repository, each with its own dependency management, build tooling, and dev server, so that changes to one workspace don't break the other and teams can work independently.

**Why this priority**: This is the foundational restructure that enables all other work. Without the monorepo structure, the web frontend cannot be added.

**Independent Test**: Clone the repo, navigate to the API workspace at `apps/api/`, and run the existing Go API with no changes to its behavior. Navigate to the web workspace at `apps/web/` and run the Next.js dev server independently. Both workspaces build and run without interfering with each other.

**Acceptance Scenarios**:

1. **Given** the repository has been restructured, **When** a developer clones it fresh, **Then** the top-level README explains the monorepo layout and how to work with each workspace.
2. **Given** the API workspace exists at `apps/api/`, **When** a developer runs the API build and test commands from within that workspace, **Then** all existing tests pass and the API starts successfully with no behavior changes.
3. **Given** the web workspace exists at `apps/web/`, **When** a developer runs the web dev server, **Then** a Next.js application starts and serves pages on its configured port.
4. **Given** both workspaces exist, **When** a developer modifies files in the web workspace, **Then** the API workspace is unaffected and vice versa.

---

### User Story 2 - Web Frontend Foundation (Priority: P2)

As a developer, I can start building the web frontend on a properly configured Next.js + React + TypeScript project, so that the frontend has a solid starting point with modern tooling and conventions.

**Why this priority**: Once the monorepo structure is in place, the web workspace needs a functional Next.js project scaffolded with the right configuration to begin feature development.

**Independent Test**: Navigate to `apps/web/`, install dependencies, start the dev server, and see a running Next.js application with TypeScript compilation, linting, and a basic page rendering.

**Acceptance Scenarios**:

1. **Given** the web workspace is initialized, **When** a developer installs dependencies and starts the dev server, **Then** a Next.js application compiles and serves at least one page without errors.
2. **Given** the web workspace has TypeScript configured, **When** a developer writes a `.tsx` file with type errors, **Then** the type checker reports the errors before build.
3. **Given** the web workspace has linting configured, **When** a developer runs the lint command, **Then** code style issues are reported.
4. **Given** the web workspace has Vitest configured, **When** a developer runs the test command, **Then** any existing tests execute and the test runner exits cleanly.

---

### User Story 3 - Shared Development Workflow (Priority: P3)

As a developer, I can use top-level Taskfile commands to build, test, and lint both workspaces at once, so that CI and local development have a single entry point for repo-wide quality checks.

**Why this priority**: After both workspaces are independently functional, developers need convenience commands to operate across the whole repo without switching directories.

**Independent Test**: Run a top-level Taskfile command from the repository root and verify it invokes the corresponding command in both the API and web workspaces.

**Acceptance Scenarios**:

1. **Given** the repository root has a Taskfile configured, **When** a developer runs the top-level test command, **Then** both the API tests (Go) and web tests (Vitest) execute and results are reported.
2. **Given** the repository root has a Taskfile configured, **When** a developer runs the top-level lint command, **Then** both the API linter (golangci-lint) and web linter (ESLint) execute.
3. **Given** the repository root has a Taskfile configured, **When** one workspace's tests fail, **Then** the failure is clearly attributed to the correct workspace.

---

### User Story 4 - Docker Compose Multi-Service (Priority: P4)

As a developer, I can run the full stack locally using Docker Compose with the API, web frontend, database, and Redis all starting together, so that I can develop and test against a complete environment.

**Why this priority**: After the individual workspaces are functional, developers need a way to run everything together for integration testing and local development.

**Independent Test**: Run `docker compose up` from the repository root and verify that the API, web, database, and Redis containers all start and the web frontend can reach the API.

**Acceptance Scenarios**:

1. **Given** Docker Compose is configured for the monorepo, **When** a developer runs `docker compose up`, **Then** all services start — API (from `apps/api/`), web (from `apps/web/`), PostgreSQL, and Redis.
2. **Given** all services are running, **When** the web frontend makes a request to the API, **Then** the request succeeds (the services can communicate).
3. **Given** a developer modifies API code in `apps/api/`, **When** the API container detects the change, **Then** the API reloads automatically (hot reload in dev mode).
4. **Given** a developer modifies web code in `apps/web/`, **When** the web container detects the change, **Then** the Next.js dev server hot reloads the page.

---

### Edge Cases

- Each workspace MUST be independently buildable and runnable; a developer working only on the API does not need Node.js installed, and vice versa. Top-level orchestration commands that span both workspaces require both toolchains.
- Each workspace maintains its own `.env` file (`apps/api/.env`, `apps/web/.env`). A root-level `.env.example` documents all variables across both workspaces for reference.
- The Go module path (`arsen`) is preserved by keeping `go.mod` inside `apps/api/`. All existing Go import paths continue to work unchanged within the `apps/api/` workspace.
- The web workspace reads the API base URL from an environment variable (e.g., `NEXT_PUBLIC_API_URL`), allowing configuration per environment without hardcoded values.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Repository MUST be organized as a monorepo with workspaces at `apps/api/` (Go backend) and `apps/web/` (Next.js frontend).
- **FR-002**: The API workspace at `apps/api/` MUST contain all existing Go source code, configuration, migrations, and tests with no functional changes to the API's behavior.
- **FR-003**: The web workspace at `apps/web/` MUST be a properly initialized Next.js project using React and TypeScript.
- **FR-004**: Each workspace MUST have its own dependency management (Go modules for API, pnpm for web).
- **FR-005**: Each workspace MUST be independently buildable, testable, and runnable without requiring the other workspace to be present or built first.
- **FR-006**: The repository MUST have a root-level Taskfile that orchestrates build, test, and lint commands across both workspaces.
- **FR-007**: Docker Compose MUST be updated to support running both the API and web services together in development mode with hot reload, with build contexts pointing to `apps/api/` and `apps/web/` respectively.
- **FR-008**: The existing Go module path (`arsen`) and import structure MUST continue to work after the restructure — `go.mod` resides in `apps/api/` and no import paths change.
- **FR-009**: The web workspace MUST have TypeScript strict mode enabled.
- **FR-010**: The web workspace MUST have ESLint and Prettier configured for code quality.
- **FR-011**: The repository root MUST have an updated README explaining the monorepo layout, workspace locations, and how to get started with each.
- **FR-012**: The web workspace MUST use Vitest with React Testing Library as its testing framework, configured and ready for use.

### Key Entities

- **API Workspace** (`apps/api/`): The Go backend containing all existing auth, user, and health features, plus infrastructure packages.
- **Web Workspace** (`apps/web/`): The Next.js frontend application that will eventually consume the API.
- **Root Orchestration**: Top-level Taskfile and Docker Compose configuration that coordinates across both workspaces.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can clone the repository and have the API running within 5 minutes using existing instructions.
- **SC-002**: A developer can clone the repository and have the web dev server running within 5 minutes.
- **SC-003**: All existing API tests pass without modification after the restructure.
- **SC-004**: A single Taskfile command from the repository root runs tests across both workspaces.
- **SC-005**: Docker Compose brings up the full stack (API + web + Postgres + Redis) in a single command.
- **SC-006**: Modifying a file in one workspace does not trigger rebuilds or errors in the other workspace.

## Assumptions

- The Go module path (`arsen`) will be preserved by keeping `go.mod` inside `apps/api/`; the restructure will not require changing import paths across existing Go source files.
- The web workspace will use the App Router (Next.js 13+ convention) rather than the Pages Router.
- pnpm will be used as the package manager for the web workspace.
- The web workspace starts as a minimal scaffold — actual feature pages (login, registration, dashboard) are out of scope for this spec and will be separate features.
- Each workspace has its own `.env` file; a root-level `.env.example` documents all variables for reference.
- CI/CD pipeline configuration is out of scope for this spec — it will be addressed separately once the monorepo structure is stable.
- Taskfile is the sole build orchestration tool; no Turborepo or Nx.
