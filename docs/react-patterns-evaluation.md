# React Patterns Exercises — Review

**Intern:** Sara Stojilkova
**Repository:** https://github.com/Sara-Stojilkova/GoInvoicing
**PRs:** #13 through #20
**Date:** 15 April 2026
**Reviewer:** Stefan

---

## Overview

Hi Sara, this is the review of your React frontend work — the 8 exercises from the React Patterns guide. I've gone through every TypeScript file, every test, the git history, all 8 PRs, and the full stack integration. Read through the full review below.

---

## This is a pass. Solid work.

You completed all 8 exercises with consistent TDD and clean process. Here's the snapshot:

| | Go work (round 2) | React exercises |
|---|---|---|
| PRs | 2 (issues #8 + #9) | 8 PRs (#13-#20), one per exercise |
| TDD evidence | Consistent test-first commits | Same — test commits before implementation throughout |
| Tests | 254 Go tests | ~160 frontend test cases across 11 test files |
| useEffect for fetching | N/A | **Zero** — all through TanStack Query |
| Conventional commits | All correct | All correct |

---

## Exercise-by-Exercise

### Exercise 1: API Client (PR #13) — Pass

Your API client is well structured — you split it across multiple files (`error.ts`, `client.ts`, `tasks.ts`, `users.ts`, `agencies.ts`) which keeps each file focused.

- Generic `request<T>` wrapper with JSON parsing, content-type checking, and error handling
- `ApiError` class with `status` and `message`
- Typed functions for every endpoint — tasks, users, and agencies
- TypeScript interfaces correctly mirror Go structs — including `string | null` for Go pointer fields and union types for status/priority/role

Good detail: your Go `Task` struct uses `json:"id"` style tags (snake_case). Your TypeScript `Task` interface uses matching snake_case field names (`agency_id`, `assignee_id`, `due_date`). The types align correctly.

**Tests:** 61 test cases covering ApiError, the request wrapper, and every API function. TDD visible — test commits before implementation commits.

### Exercise 2: Query & Mutation Hooks (PR #14) — Pass

- `useTasks(agencyId)`, `useTask(taskId, agencyId)`, `useCreateTask(agencyId)`, `useCompleteTask(agencyId)` plus `useUsers(agencyId)`
- All use TanStack Query — zero `useEffect` in the entire frontend
- Cache invalidation on mutation success — scoped to the correct agency ID
- Conditional fetching: `useTask` uses `enabled: taskId !== null`
- QueryClientProvider set up in `main.tsx`

Strong test detail: you test that cache invalidation happens for the right agency AND that other agencies' caches are NOT invalidated. This shows you understand how query key scoping works — not just that invalidation happens, but that it's targeted.

**Tests:** 22 hook tests with `renderHook`.

### Exercise 3: Task List Page (PR #15) — Pass

- Loading state with spinner and `role="status"`
- Error state with message and a "Retry" button that calls `refetch()`
- Empty state: "No tasks found."
- `StatusBadge` component with color mapping (todo=gray, in_progress=yellow, done=green)
- Complete button disabled while pending, shows spinner during mutation

Good accessibility: `role="status"` for loading, `role="alert"` for errors.

**Tests:** 27 component tests covering all three phases plus interaction.

### Exercise 4: Create Task Form (PR #16) — Pass

- Controlled inputs for title, priority, description, assignee, due date
- Client-side validation for required title
- Submit button disabled while pending
- Error message displayed on mutation failure
- All fields reset on success

You went beyond the exercise by adding an assignee dropdown (fetched via `useUsers` hook) and modifying the Go backend to accept the new fields. That's real full-stack thinking — adjusting both sides to make the feature work.

**Tests:** 17 test cases including validation, pending state, error display, field reset, and assignee selection.

### Exercise 5: Filtering & Summary Cards (PR #17) — Pass (one miss)

- `useState<StatusFilter>` for the filter
- `useMemo` for the filtered task list — correct
- `useMemo` for summary counts from the **full** task list — correct (counts always reflect the complete dataset)
- Clickable summary cards with `aria-pressed` for accessibility
- Active card gets `summary-card--active` class

**One thing missing:** the exercise says "clicking the active filter card should clear the filter (show all)." Your current code re-sets the same filter when you click an already-active card. The toggle behavior would be one line: `onFilterChange(activeFilter === value ? "all" : value)`. This is the only functional gap across all 8 exercises.

**Tests:** 21 tests across SummaryCards and TaskListPage covering counts, filter interaction, and active state.

### Exercise 6: Routing & Detail Page (PR #18) — Pass

- React Router with `BrowserRouter`, `Routes`, `Route`, `Link`
- Both pages lazy-loaded with `lazy()` + `Suspense` fallback
- `useParams<{ taskId: string }>()` on the detail page
- Task titles in the list are `<Link>` elements navigating to `/tasks/{id}`
- "Back to list" link on the detail page
- Detail page shows all fields: title, status, priority, ID, agency, assignee, dates, description

Nice detail: `TaskDetailPage` is a named export, and you correctly handled the `lazy()` requirement for default exports using `.then(m => ({ default: m.TaskDetailPage }))`. That shows you understand how lazy loading works under the hood.

**Tests:** 24 test cases — the most thorough detail page tests. Covers loading, error, not found, navigation, all required fields, and all optional fields.

### Exercise 7: Tests (PR #19) — Pass

- Vitest configured with jsdom environment
- Testing Library with `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`
- Shared test infrastructure: `createTestQueryClient()`, `createWrapper()`, `createRouterWrapper()`, `createPageWrapper()`
- Hook tests with `renderHook` + `waitFor`
- Component tests for every component
- Table-driven tests with `it.each` for StatusBadge

Your test-to-code ratio is about 1.6:1 (930 lines of tests to 576 lines of production code). That's healthy.

### Exercise 8: Full Stack Integration (PR #20) — Pass

- CORS middleware on Go backend with configurable `CORS_ORIGIN` env var (defaults to `http://localhost:5173`), handles OPTIONS preflight
- Vite proxy forwarding `/api` to `http://localhost:8080`

---

## Process

Your process is consistent and clean:

- **8 PRs**, one per exercise, all on feature branches
- **Conventional commits** on every commit — `feat:`, `test:`, `fix:`, `refactor:` throughout
- **TDD clearly visible** in the commit history:
  - `test: wrote tests for the new ApiError class` → then `feat: implemented ApiError`
  - `test: a hook that fetches the task list` → then `feat: added a hook that fetches task list`
  - `test: TaskListPage and its phases` → then `feat: created a page that uses the query hook`
  - `test: CreateTaskForm component` → then `feat: implemented CreateTaskForm`
  - This pattern is consistent across all 8 PRs
- **Zero `useEffect` for data fetching** — the one rule, followed completely
- **Types correctly mirror the Go backend** — field names, nullable types, union types all aligned

---

## Things to improve

These are minor — they won't block you from real work:

### 1. Missing filter toggle behavior

Exercise 5 asked for clicking an active filter card to clear the filter back to "all." Your code doesn't toggle — clicking the active "Todo" card keeps the filter on "Todo." The fix is one line in the click handler:

```
onFilterChange(activeFilter === value ? "all" : value)
```

### 2. Hardcoded agency ID

`App.tsx` has a hardcoded UUID for the agency. This is fine for the exercise, but in the real project this would come from authentication context. Something to be aware of when you start working on the compliance tracker.

### 3. Inconsistent default exports

`TaskListPage` has both a named export and `export default`. `TaskDetailPage` only has a named export (requiring the `.then()` workaround in `lazy()`). Pick one pattern and be consistent — either always use default exports for pages, or always use named exports with the `.then()` adapter.

### 4. Test environment setup

Your `vite.config.ts` sets `environment: 'node'` globally and individual test files opt into jsdom with `// @vitest-environment jsdom`. This works but is unusual — most React projects set jsdom globally since component tests are the majority. The per-file pragma is easy to forget.

### 5. One typo

Commit `c3a0716` says "fix: brokent tests" — should be "broken."

---

## Where you are now

You've completed the Go code patterns, the task management API with full test coverage, and the React frontend exercises. You understand:

- Backend: domain entities, repository interfaces, service layer, DI, HTTP handlers, multi-tenancy, TDD
- Frontend: typed API clients, TanStack Query, component composition, routing, filtering, testing
- Process: TDD, conventional commits, feature branches, PRs, code review

You're ready for real tasks on the compliance tracker. The patterns you practiced are the same ones you'll find in `frontend/src/hooks/useProperties.ts` and `frontend/src/pages/DashboardPage.tsx` in our production codebase.
