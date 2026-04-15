# Task Management API & Issue Fixes — Review

**Intern:** Sara Stojilkova
**Repository:** https://github.com/Sara-Stojilkova/GoInvoicing
**Date:** 6 April 2026
**Reviewer:** Stefan

---

## Overview

Hi Sara, this is the review of your work on issue #8 (fixes from the first evaluation) and issue #9 (task management API). I've gone through the full repo: every Go file, every test, the git history, all PRs, and ran the test suite. Read through the full review, then see the next assignment at the bottom.

---

## The improvement is clear

You took every piece of feedback from the first evaluation and applied it. Here's what changed:

| | GoInvoicing (first eval) | After issues #8 + #9 |
|---|---|---|
| Conventional commits | 0 out of 13 | All commits use `feat:`, `fix:`, `test:`, `refactor:` |
| TDD evidence | None — tests and code always in same commit | Tests committed before implementation across multiple features |
| Handler tests | None | 32+ table-driven cases across task, user, agency, invoice handlers |
| `MarkAsPaid` on domain | Missing — logic only in service | Added to `Invoice` struct with proper error wrapping |
| Broken import | `invoice_service_test.go` importing deleted `postgres` package | Fixed — now imports `repositories/memory` |
| Missing domain tests | `DaysUntilDue`, `CalculateLateFee`, `SummarizeInvoices` untested | All added — 11, 7, and 6 table-driven cases respectively |
| Test suite | Wasn't compiling | **254 tests, all passing** |

---

## Issue #8: Fixes — Complete

All three requirements from the first evaluation were addressed:

### 1. Broken import fixed

Commit `d552fab` — changed `postgres.NewInvoiceRepo()` to `memory.NewInvoiceRepo()` in `invoice_service_test.go`. Tests compile and pass.

### 2. `MarkAsPaid` added to the domain entity

Commit `32a532e` — the method is on the `Invoice` struct where it belongs:

```go
func (i *Invoice) MarkAsPaid(now time.Time) error {
    if i.IsPaid() {
        return fmt.Errorf("invoice %s: %w", i.ID, apperrors.ErrConflict)
    }
    i.PaidAt = &now
    i.Status = "paid"
    return nil
}
```

The service now calls `invoice.MarkAsPaid(now)` instead of handling the validation itself. This is the right design — the entity protects its own invariants.

### 3. All missing tests added

| Function | Test Cases | Status |
|----------|-----------|--------|
| `DaysUntilDue` | 11 table-driven | Pass |
| `CalculateLateFee` | 7 table-driven | Pass |
| `SummarizeInvoices` | 6 table-driven | Pass |
| `MarkAsPaid` (domain) | 3 table-driven | Pass |
| Invoice handler tests (`httptest`) | 17 across 4 test functions | Pass |

---

## Issue #9: Task Management API — Complete

This is a substantial piece of work. You didn't just build a task API — you built a multi-entity system with agencies, users, and tasks, including access control. 24 commits over the feature branch.

### Architecture

All patterns from the code patterns exercises are correctly implemented:

- **Three domain entities** — `Task`, `User`, `Agency`. All pure — no `net/http`, no `database/sql`, no framework imports. Clean.
- **Domain methods with validation** — `Task.Complete(now)` returns `ErrConflict` if already done. `Task.SetInProgress()` returns error if already in progress. `Task.IsAccessibleBy(agencyID)` enforces multi-tenancy. These are entity-level invariants, exactly where they belong.
- **Repository interfaces** for all three entities with in-memory implementations using `sync.RWMutex`.
- **Service layer depends on interfaces** via constructor injection. All three services follow the same pattern: `NewTaskService(repo repositories.TaskRepository)`.
- **Sentinel errors** — `ErrNotFound`, `ErrConflict`, `ErrForbidden` in `apperrors/`, wrapped with `%w` throughout repository and service layers, checked with `errors.Is()` in handlers.
- **Thin HTTP handlers** — parse request, call service, write response. Error mapping: `ErrNotFound` → 404, `ErrForbidden` → 403, `ErrConflict` → 409.
- **Graceful shutdown** in `cmd/server/main.go` with signal handling and context timeout.
- **Chi router** with clean route grouping under `/api/tasks`, `/api/users`, `/api/agencies`.

### Multi-tenancy

You added agency-based access control that wasn't in the assignment. Tasks belong to an agency, and the service layer checks that users can only access tasks within their agency:

```go
func (t Task) IsAccessibleBy(userAgencyID uuid.UUID) bool {
    return t.AgencyID == userAgencyID
}
```

This shows you're thinking about how the real compliance tracker works — it has the same agency-scoping pattern.

### Tests

**254 tests, all passing.** Coverage across every layer:

| Layer | Test Files | Test Functions | Style |
|-------|-----------|----------------|-------|
| Domain (task) | `task_test.go` | 12 functions, 54+ cases | Table-driven |
| Domain (user) | `user_test.go` | 2 functions | Table-driven |
| Domain (invoice) | `invoice_test.go`, `invoice_rules_test.go` | 6 functions | Table-driven |
| Service (task) | `task_service_test.go` | 7 functions, 22 cases | Table-driven |
| Service (user) | `user_service_test.go` | 3 functions | Table-driven |
| Service (agency) | `agency_service_test.go` | 3 functions | Table-driven |
| Handler (task) | `task_handler_test.go` | 6 functions, 32 cases | Table-driven + `httptest` |
| Handler (user) | `user_handler_test.go` | 3 functions | Table-driven + `httptest` |
| Handler (agency) | `agency_handler_test.go` | 3 functions | Table-driven + `httptest` |

Handler tests use `httptest.NewRecorder()` — the gap from the first evaluation is fully addressed.

### TDD in the git history

The commit ordering shows test-first workflow across multiple features:

- `test: new user and task domain entities` → then `feat: added user and task domains`
- `test: new services to implement (task, user, agency)` → then `feat: task, user and agency services`
- `test: reopening and setting tasks in progress in domain` → then `feat: implementation of StartProgress and Reopen methods`
- `test: SetInProgress method in task_service` → then `feat: implemented SetInProgress service method`
- `test: agency and user handlers` → then `feat: implemented user and agency handlers`

This pattern is consistent across the entire feature branch.

### Conventional commits

All commits follow the format — `feat:`, `fix:`, `test:`, `refactor:`. No misses this time.

---

## Things to note

These aren't blockers — they're observations for your growth:

### 1. Module name is still `backend`

`go.mod` says `module backend`. This works, but it's generic. For a multi-project repo it's fine since the module is inside the `backend/` directory, but if this were a standalone project you'd want something more specific like `module go-invoicing` or `module github.com/Sara-Stojilkova/GoInvoicing`.

### 2. Go version 1.18

`go.mod` specifies `go 1.18` which is from 2022. The current version is 1.22+. Not a functional issue, but worth updating when you get the chance.

### 3. Refactoring mid-feature

Commits like `refactor: separated domains in different folders` and `refactor: moved invoice service to a separate folder` show you restructured the project while building the task API. This is fine in a practice repo, but on the real project, keep refactoring PRs separate from feature PRs. It makes review easier — the reviewer can see "this PR only moves files" vs "this PR adds new behavior."

---

## Next Assignment: Present Your Work

You've proven you can write the code, follow the process, and respond to feedback. The next skill to build is **explaining what you built and why**.

Engineering is not just writing code — it's communicating your decisions to other people. On the real project, you'll need to explain your PRs in review, discuss architecture choices in team syncs, and demo features to stakeholders. This assignment is practice for all of that.

### What to prepare

**A 15-20 minute presentation + live demo of your task management API.**

You'll present this to the team over a call. The audience is Stefan and the other interns. Assume they haven't seen your code — walk them through it.

### Structure

**Part 1: Architecture (5-7 minutes)**

Explain the structure of your project. Not "here's every file" — explain the *layers* and *why* they're separated the way they are. Cover:

- What's in the domain layer and why it doesn't import any framework packages
- What problem the repository interface solves — why not just call the database directly from the service?
- How dependency injection works in your `main.go` — how the pieces connect
- The multi-tenancy design — why tasks are scoped to agencies and how the access control works
- One decision you made that you're proud of, and one thing you'd do differently

Use a diagram if it helps. A simple boxes-and-arrows sketch is fine — it doesn't need to be fancy.

**Part 2: Live demo (5-7 minutes)**

Run the server and walk through the API using curl or a tool like Postman/Insomnia:

- Create an agency, create a user in that agency, create a task
- Assign the task to a user, set it in progress, complete it
- Show filtering — list tasks by agency, list overdue tasks
- Show an error case: try to access a task from a different agency (403), try to complete an already-completed task
- Show what happens when you create a task with an empty title

Don't just show happy paths. Show how the system handles mistakes — that's where the architecture proves itself.

**Part 3: Code walkthrough (3-5 minutes)**

Pick one feature and trace it through the full stack — from the HTTP request to the stored data. For example: "What happens when someone calls `POST /tasks/{id}/assign`?"

Walk through: handler parses the request and extracts the task ID and assignee ID → calls the service → service gets the task from the repo, checks agency access, calls `task.Assign()` on the domain entity → calls repo.Update to save it. Show the actual code for each step.

**Part 4: What you learned (2-3 minutes)**

Talk about your experience across both rounds of work:

- What changed between the first submission and the issue #8 + #9 work
- What was the hardest part of building the task management API
- What you'd do differently if you started the project from scratch
- How the multi-tenancy design came about — was it planned from the start or did it emerge?

### How to prepare

- **Practice out loud at least twice.** Presenting to yourself (or a mirror, or a stuffed animal) feels weird, but it's the fastest way to find the parts where you stumble. If you can say it smoothly alone, you can say it smoothly in the call.
- **Keep slides minimal.** A few slides with diagrams and key points. The code and the demo are the main event, not the slides. Don't read from slides — use them as a visual aid.
- **It's okay to say "I don't know."** If someone asks a question and you're not sure of the answer, say that. "I'm not sure, let me check" is always better than guessing.
- **Slow down.** When people are nervous they talk fast. Consciously slow your pace. Pauses are fine — they give the audience time to absorb what you said.
- **Have your terminal ready.** Before the call, make sure the server starts, have your curl commands ready in a text file so you can copy-paste during the demo. Nothing kills a demo like typing a long curl command live and getting a typo.

### When

Coordinate a time with Stefan. Aim for sometime next week — you need a few days to prepare, but don't over-prepare. The goal is to communicate clearly, not to give a perfect performance.

### Why this matters

On the compliance tracker, you'll be opening PRs that other people review. You'll be in team syncs where you explain what you're working on. Eventually you'll demo features. The ability to explain your technical decisions clearly and confidently is as important as the code itself. This presentation is practice for all of that — in a safe environment where the only goal is to get more comfortable.
