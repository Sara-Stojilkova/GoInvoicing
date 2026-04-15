// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { createTestQueryClient } from "../test/testQueryClient";
import { TaskDetailPage } from "./TaskDetailPage";
import * as tasksApi from "../api/tasks";
import type { Task } from "../types/api";

const agencyId   = "a1b2c3d4-0000-0000-0000-000000000001";
const taskId     = "c3d4e5f6-0000-0000-0000-000000000003";
const assigneeId = "b2c3d4e5-0000-0000-0000-000000000002";

const fullTask: Task = {
  id:           taskId,
  title:        "Fix login bug",
  description:  "Users cannot log in with SSO enabled",
  status:       "in_progress",
  priority:     "high",
  agency_id:    agencyId,
  assignee_id:  assigneeId,
  created_at:   "2024-01-15T09:00:00Z",
  due_date:     "2024-02-01T00:00:00Z",
  completed_at: null,
};

const minimalTask: Task = {
  id:           taskId,
  title:        "Minimal task",
  description:  null,
  status:       "todo",
  priority:     "low",
  agency_id:    agencyId,
  assignee_id:  null,
  created_at:   "2024-01-15T09:00:00Z",
  due_date:     null,
  completed_at: null,
};

const completedTask: Task = {
  ...fullTask,
  status:       "done",
  completed_at: "2024-01-20T14:30:00Z",
};

function renderPage(id = taskId) {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/tasks/${id}`]}>
        <Routes>
          <Route path="/tasks/:taskId" element={<TaskDetailPage agencyId={agencyId} />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => vi.restoreAllMocks());

describe("TaskDetailPage", () => {
  describe("loading state", () => {
    it("renders a loading indicator while fetching", () => {
      vi.spyOn(tasksApi, "getTask").mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("does not render task content while loading", () => {
      vi.spyOn(tasksApi, "getTask").mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.queryByText("Fix login bug")).not.toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders an error message when the request fails", async () => {
      vi.spyOn(tasksApi, "getTask").mockRejectedValue(new Error("network error"));
      renderPage();
      await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
    });

    it("shows a not found message on 404", async () => {
      vi.spyOn(tasksApi, "getTask").mockRejectedValue(
        Object.assign(new Error("not found"), { status: 404 })
      );
      renderPage();
      await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
      expect(screen.getByText(/not found/i)).toBeInTheDocument();
    });

    it("shows a back link on the error state", async () => {
      vi.spyOn(tasksApi, "getTask").mockRejectedValue(new Error("network error"));
      renderPage();
      await waitFor(() => screen.getByRole("alert"));
      expect(screen.getByRole("link", { name: /back/i })).toBeInTheDocument();
    });
  });

  describe("navigation", () => {
    it("renders a back link to the list", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("link", { name: /back/i })).toHaveAttribute("href", "/");
    });
  });

  describe("success state — required fields", () => {
    beforeEach(() => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
    });

    it("reads the task id from the URL and fetches the right task", async () => {
      const spy = vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      renderPage(taskId);
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(spy).toHaveBeenCalledWith(taskId, agencyId);
    });

    it("renders the task title", async () => {
      renderPage();
      await waitFor(() => expect(screen.getByText("Fix login bug")).toBeInTheDocument());
    });

    it("renders the task id", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(taskId)).toBeInTheDocument();
    });

    it("renders the status", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/in.progress/i)).toBeInTheDocument();
    });

    it("renders the priority", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/high/i)).toBeInTheDocument();
    });

    it("renders the agency id", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(agencyId)).toBeInTheDocument();
    });

    it("renders the created_at date", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/jan.*15.*2024|2024.*01.*15/i)).toBeInTheDocument();
    });
  });

  describe("success state — optional fields", () => {
    it("renders the description when present", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText("Users cannot log in with SSO enabled")).toBeInTheDocument();
    });

    it("does not render a description section when description is null", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      expect(screen.queryByText(/description/i)).not.toBeInTheDocument();
    });

    it("renders the assignee_id when present", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(assigneeId)).toBeInTheDocument();
    });

    it("shows unassigned when assignee_id is null", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      expect(screen.getByText(/unassigned/i)).toBeInTheDocument();
    });

    it("renders the due date when present", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/feb.*1.*2024|2024.*02.*01/i)).toBeInTheDocument();
    });

    it("shows no due date when due_date is null", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      expect(screen.getByText(/no due date/i)).toBeInTheDocument();
    });

    it("renders the completed_at date when present", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(completedTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/jan.*20.*2024|2024.*01.*20/i)).toBeInTheDocument();
    });

    it("does not render a completed section when task is not done", async () => {
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.queryByText(/completed/i)).not.toBeInTheDocument();
    });
  });
});
