// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { TaskListPage } from "./TaskListPage";
import * as tasksApi from "../api/tasks";
import type { Task } from "../types/api";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";

const tasks: Task[] = [
  {
    id: "c3d4e5f6-0000-0000-0000-000000000003",
    title: "Fix login bug",
    description: null,
    status: "todo",
    priority: "high",
    agency_id: agencyId,
    assignee_id: null,
    created_at: "2024-01-01T00:00:00Z",
    due_date: null,
    completed_at: null,
  },
  {
    id: "d4e5f6a7-0000-0000-0000-000000000004",
    title: "Write docs",
    description: null,
    status: "in_progress",
    priority: "low",
    agency_id: agencyId,
    assignee_id: null,
    created_at: "2024-01-02T00:00:00Z",
    due_date: null,
    completed_at: null,
  },
];

function renderPage(agencyId: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    createElement(QueryClientProvider, { client: queryClient },
      createElement(TaskListPage, { agencyId })
    )
  );
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("TaskListPage", () => {
  describe("loading state", () => {
    it("renders a loading indicator while fetching", () => {
      vi.spyOn(tasksApi, "listTasks").mockReturnValue(new Promise(() => {}));

      renderPage(agencyId);

      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("does not render any task titles while loading", () => {
      vi.spyOn(tasksApi, "listTasks").mockReturnValue(new Promise(() => {}));

      renderPage(agencyId);

      expect(screen.queryByText("Fix login bug")).not.toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders an error message when the request fails", async () => {
      vi.spyOn(tasksApi, "listTasks").mockRejectedValue(new Error("network error"));

      renderPage(agencyId);

      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument()
      );
    });

    it("shows a retry button on error", async () => {
      vi.spyOn(tasksApi, "listTasks").mockRejectedValue(new Error("network error"));

      renderPage(agencyId);

      await waitFor(() =>
        expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument()
      );
    });

    it("retries the request when the retry button is clicked", async () => {
      const spy = vi.spyOn(tasksApi, "listTasks")
        .mockRejectedValueOnce(new Error("network error"))
        .mockResolvedValue(tasks);

      renderPage(agencyId);

      await waitFor(() => screen.getByRole("button", { name: /retry/i }));
      await userEvent.click(screen.getByRole("button", { name: /retry/i }));

      await waitFor(() => expect(spy).toHaveBeenCalledTimes(2));
    });

    it("shows tasks after a successful retry", async () => {
      vi.spyOn(tasksApi, "listTasks")
        .mockRejectedValueOnce(new Error("network error"))
        .mockResolvedValue(tasks);

      renderPage(agencyId);

      await waitFor(() => screen.getByRole("button", { name: /retry/i }));
      await userEvent.click(screen.getByRole("button", { name: /retry/i }));

      await waitFor(() =>
        expect(screen.getByText("Fix login bug")).toBeInTheDocument()
      );
    });
  });

  describe("success state", () => {
    it("renders a list of tasks on success", async () => {
      vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

      renderPage(agencyId);

      await waitFor(() =>
        expect(screen.getByText("Fix login bug")).toBeInTheDocument()
      );
      expect(screen.getByText("Write docs")).toBeInTheDocument();
    });

    it("renders each task's status", async () => {
      vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

      renderPage(agencyId);

      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/todo/i)).toBeInTheDocument();
      expect(screen.getByText(/in.progress/i)).toBeInTheDocument();
    });

    it("renders each task's priority", async () => {
      vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

      renderPage(agencyId);

      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/high/i)).toBeInTheDocument();
      expect(screen.getByText(/low/i)).toBeInTheDocument();
    });

    it("renders an empty state message when there are no tasks", async () => {
      vi.spyOn(tasksApi, "listTasks").mockResolvedValue([]);

      renderPage(agencyId);

      await waitFor(() =>
        expect(screen.getByText(/no tasks/i)).toBeInTheDocument()
      );
    });

    it("does not render the loading indicator after tasks load", async () => {
      vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

      renderPage(agencyId);

      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });
});
