// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SummaryCards } from "./SummaryCards";
import type { StatusFilter } from "./SummaryCards";
import type { Task } from "../types/api";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";

const tasks: Task[] = [
  { id: "1", title: "A", description: null, status: "todo",        priority: "high",   agency_id: agencyId, assignee_id: null, created_at: "2024-01-01T00:00:00Z", due_date: null, completed_at: null },
  { id: "2", title: "B", description: null, status: "todo",        priority: "medium", agency_id: agencyId, assignee_id: null, created_at: "2024-01-01T00:00:00Z", due_date: null, completed_at: null },
  { id: "3", title: "C", description: null, status: "in_progress", priority: "low",    agency_id: agencyId, assignee_id: null, created_at: "2024-01-01T00:00:00Z", due_date: null, completed_at: null },
  { id: "4", title: "D", description: null, status: "done",        priority: "high",   agency_id: agencyId, assignee_id: null, created_at: "2024-01-01T00:00:00Z", due_date: null, completed_at: null },
  { id: "5", title: "E", description: null, status: "done",        priority: "low",    agency_id: agencyId, assignee_id: null, created_at: "2024-01-01T00:00:00Z", due_date: null, completed_at: null },
];

function renderCards(activeFilter: StatusFilter = "all", onFilterChange = vi.fn()) {
  return { onFilterChange, ...render(<SummaryCards tasks={tasks} activeFilter={activeFilter} onFilterChange={onFilterChange} />) };
}

beforeEach(() => vi.restoreAllMocks());

describe("SummaryCards", () => {
  describe("counts", () => {
    it("shows the total task count", () => {
      renderCards();
      expect(screen.getByRole("button", { name: /all/i })).toHaveTextContent("5");
    });

    it("shows the todo count", () => {
      renderCards();
      expect(screen.getByRole("button", { name: /todo/i })).toHaveTextContent("2");
    });

    it("shows the in-progress count", () => {
      renderCards();
      expect(screen.getByRole("button", { name: /in.progress/i })).toHaveTextContent("1");
    });

    it("shows the done count", () => {
      renderCards();
      expect(screen.getByRole("button", { name: /done/i })).toHaveTextContent("2");
    });

    it("updates counts when tasks change", () => {
      const { rerender } = render(
        <SummaryCards tasks={tasks} activeFilter="all" onFilterChange={vi.fn()} />
      );
      rerender(
        <SummaryCards tasks={tasks.slice(0, 1)} activeFilter="all" onFilterChange={vi.fn()} />
      );
      expect(screen.getByRole("button", { name: /all/i })).toHaveTextContent("1");
    });
  });

  describe("filtering", () => {
    it("calls onFilterChange with 'todo' when the todo card is clicked", async () => {
      const { onFilterChange } = renderCards();
      await userEvent.click(screen.getByRole("button", { name: /todo/i }));
      expect(onFilterChange).toHaveBeenCalledWith("todo");
    });

    it("calls onFilterChange with 'in_progress' when the in-progress card is clicked", async () => {
      const { onFilterChange } = renderCards();
      await userEvent.click(screen.getByRole("button", { name: /in.progress/i }));
      expect(onFilterChange).toHaveBeenCalledWith("in_progress");
    });

    it("calls onFilterChange with 'done' when the done card is clicked", async () => {
      const { onFilterChange } = renderCards();
      await userEvent.click(screen.getByRole("button", { name: /done/i }));
      expect(onFilterChange).toHaveBeenCalledWith("done");
    });

    it("calls onFilterChange with 'all' when the all card is clicked", async () => {
      const { onFilterChange } = renderCards("todo");
      await userEvent.click(screen.getByRole("button", { name: /all/i }));
      expect(onFilterChange).toHaveBeenCalledWith("all");
    });
  });

  describe("toggle behaviour", () => {
    it("calls onFilterChange with 'all' when the active todo card is clicked again", async () => {
      const { onFilterChange } = renderCards("todo");
      await userEvent.click(screen.getByRole("button", { name: /todo/i }));
      expect(onFilterChange).toHaveBeenCalledWith("all");
    });

    it("calls onFilterChange with 'all' when the active in-progress card is clicked again", async () => {
      const { onFilterChange } = renderCards("in_progress");
      await userEvent.click(screen.getByRole("button", { name: /in.progress/i }));
      expect(onFilterChange).toHaveBeenCalledWith("all");
    });

    it("calls onFilterChange with 'all' when the active done card is clicked again", async () => {
      const { onFilterChange } = renderCards("done");
      await userEvent.click(screen.getByRole("button", { name: /done/i }));
      expect(onFilterChange).toHaveBeenCalledWith("all");
    });

    it("calls onFilterChange with 'todo' when an inactive todo card is clicked", async () => {
      const { onFilterChange } = renderCards("done");
      await userEvent.click(screen.getByRole("button", { name: /todo/i }));
      expect(onFilterChange).toHaveBeenCalledWith("todo");
    });
  });

  describe("active state", () => {
    it("marks the active card with aria-pressed", () => {
      renderCards("todo");
      expect(screen.getByRole("button", { name: /todo/i })).toHaveAttribute("aria-pressed", "true");
    });

    it("does not mark inactive cards with aria-pressed", () => {
      renderCards("todo");
      expect(screen.getByRole("button", { name: /all/i })).toHaveAttribute("aria-pressed", "false");
      expect(screen.getByRole("button", { name: /in.progress/i })).toHaveAttribute("aria-pressed", "false");
      expect(screen.getByRole("button", { name: /done/i })).toHaveAttribute("aria-pressed", "false");
    });

    it("marks 'all' as active by default", () => {
      renderCards("all");
      expect(screen.getByRole("button", { name: /all/i })).toHaveAttribute("aria-pressed", "true");
    });
  });
});
