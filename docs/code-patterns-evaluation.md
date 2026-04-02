# Code Patterns Exercise — Review

**Intern:** Sara Stojilkova
**Repository:** https://github.com/Sara-Stojilkova/GoInvoicing
**Date:** 2 April 2026
**Reviewer:** Stefan

---

## Overview

Hi Sara, this is the review of your code patterns exercise submission. I've gone through the repo — every Go file, every test, the frontend, the git history, the PRs, and the project structure. Read through the full review, then see the next assignment at the bottom.

---

## What you got right

You worked through the exercises incrementally across 3 days with feature branches and PRs. That's the right approach. Here's what landed well:

- **Domain layer is clean** — `domain/invoice.go` imports only `time` and `uuid`. No `net/http`, no database packages, no framework dependencies. Domain purity is maintained throughout.
- **Business rules are pure functions** — `EvaluateInvoiceStatus`, `CalculateLateFee`, `SummarizeInvoices` all take their inputs as parameters (including `time.Time` instead of calling `time.Now()` internally). Correct.
- **Repository interface is defined** with `Create`, `GetByID`, `List`, `Update` — and the in-memory implementation uses `sync.RWMutex` for thread safety. This matches Exercise 3.
- **Service layer depends on the interface** via constructor injection (`NewInvoiceService(repo repositories.InvoiceRepository)`). The service doesn't know or care what's behind the interface. Correct pattern.
- **HTTP handlers are separated** in `api/` — they parse requests, call the service, and write responses. Error mapping is correct (404 for not found, 409 for conflict, 400 for bad input).
- **Sentinel errors** in `apperrors/` with `ErrNotFound` and `ErrConflict`, using `%w` wrapping — correct pattern.
- **Graceful shutdown** in `cmd/server/main.go` with signal handling and context timeout.
- **Frontend uses TanStack Query** — not `useEffect` for data fetching. Typed API client in `api/client.ts`, proper hooks (`useInvoices`, `useCreateInvoice`, `useMarkPaid`, `useSummary`), query invalidation on mutations.
- **Full stack integration works** — CORS middleware, Vite proxy, SummaryCards component with filter-by-status on click.
- **Feature branches and PRs** — you used `feat/` branches and opened PRs for each exercise. This is the workflow we use on the real project.
- **`EvaluateInvoiceStatus` has 9 test cases** — exceeds the 8+ requirement. Good edge case coverage (boundary conditions on dates, paid-preserves-over-overdue).

---

## What needs to be fixed

### 1. Tests are broken — the service tests won't compile

This is the most important issue. Your `invoice_service_test.go` imports `"backend/internal/repositories/postgres"` on line 10 — but that package was deleted in a later commit. If you run `go test ./internal/services/` right now, it will fail with a compilation error.

This happened because you originally had a `repositories/postgres/` directory, then deleted it in the "Removed duplicate file" commit, but didn't update the test import to point to `repositories/memory/` instead.

This is exactly the kind of thing that gets caught by running `go test ./...` before committing. Always run the full test suite after a cleanup — especially after deleting files.

### 2. Missing `MarkAsPaid` on the domain entity

Exercise 1 says:

> *"MarkAsPaid(now time.Time) error — return error if already paid"*

This method should live on the `Invoice` struct in `domain/invoice.go`. You put the "already paid" check in the service layer instead, which makes the domain entity anemic — it holds data but doesn't protect its own invariants.

Why this matters: the domain entity should be responsible for its own rules. If `MarkAsPaid` is on the entity, then any code that has an `Invoice` can safely mark it as paid and get the validation for free. If it's only in the service, then anyone who bypasses the service (a worker, a migration script, a test) can set `PaidAt` on an already-paid invoice without the guard.

### 3. No conventional commit messages

The onboarding guide and the exercise document both specify conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`. Your commits use plain descriptions:

- `"Add invoice domain entity and tests"` — should be `feat: add invoice domain entity and tests`
- `"Added rules and tests"` — should be `feat: add business rules and tests`
- `"Removed duplicate file"` — should be `fix: remove duplicate postgres repo file`

This isn't cosmetic. On the real project, conventional commits are enforced and used for changelogs. Build the habit now.

### 4. No evidence of TDD in the git history

The exercise guide says: *"Write the test first — every exercise says 'test first.'"*

In your git history, test files and implementation files always arrive in the same commit. For example, `invoice.go` and `invoice_test.go` were committed together in `3bc1f10`. There's no commit showing a failing test followed by the implementation that makes it pass.

TDD doesn't mean "write tests." It means: write the test first, see it fail, then write the code to make it pass. The git history should show this — a commit with just the test (which won't pass yet), then a commit with the implementation.

### 5. Worker exercise was skipped

The exercise for the worker pool with `recover()` was not attempted. No worker files exist in the repo.

### 6. Missing tests for several functions

- **`DaysUntilDue`** — the method exists but has no tests
- **`CalculateLateFee`** — the function exists but has no tests
- **`SummarizeInvoices`** — only tested indirectly through the service; no direct unit tests in `rules_test.go`
- **No handler tests** — the `api/` package has no test files. Handler tests using `httptest` verify that your error mapping, input parsing, and response formatting work correctly.

---

## Fixes for the current repo

Before moving on, fix the GoInvoicing repo. Each fix should be a separate commit on a branch with a PR:

1. **Fix the broken import** — update `invoice_service_test.go` to import `repositories/memory` instead of the deleted `repositories/postgres`. Run `go test ./...` and confirm it passes.
2. **Add `MarkAsPaid(now time.Time) error` to the domain entity** — move the "already paid" validation from the service into the `Invoice` struct. Update the service to call `invoice.MarkAsPaid(now)`. Add tests for it in `invoice_test.go`.
3. **Add the missing tests** — `DaysUntilDue`, `CalculateLateFee`, `SummarizeInvoices` in `domain/`, and handler tests using `httptest` in `api/`.

Use conventional commit messages for all of these.

---

## Next Assignment

You've shown you understand the architecture patterns and you followed the branching/PR workflow. Now take it further — apply the patterns to something new, tighten up the process, and fill the gaps.

**Build a task management API.**

A simple system for managing tasks — think of it as a basic project tracker. Users can create tasks, assign them, update their status, and filter by different criteria.

Use everything you learned from the code patterns exercises. Build it in a new repository. Use Claude Code — that's expected — but you are responsible for reviewing what it produces, committing in logical steps, and following the workflow.

**What we'll be evaluating:**

- Your git history — commits, branches, PRs, and whether tests come before implementation
- Conventional commit messages on every commit
- Whether your code follows the patterns from the exercises
- Whether you reviewed what Claude Code generated (no broken imports, no dead code, no disconnected pieces)
- That `go test ./...` passes at every commit

When you're done, share the repo link with Stefan.
