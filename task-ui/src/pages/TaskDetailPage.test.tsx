import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { createTestQueryClient } from "../test/testQueryClient";
import { TaskDetailPage } from "./TaskDetailPage";
import * as tasksApi from "../api/tasks";
import * as usersApi from "../api/users";
import * as agenciesApi from "../api/agencies";
import type { Task, User, Agency } from "../types/api";

const agencyId   = "a1b2c3d4-0000-0000-0000-000000000001";
const taskId     = "c3d4e5f6-0000-0000-0000-000000000003";
const assigneeId = "b2c3d4e5-0000-0000-0000-000000000002";

const agency: Agency = {
  id: agencyId,
  name: "Acme Corp",
  created_at: "2024-01-01T00:00:00Z",
};

const assigneeUser: User = {
  id: assigneeId,
  name: "Alice",
  email: "alice@acme.com",
  role: "admin",
  agency_id: agencyId,
  created_at: "2024-01-01T00:00:00Z",
};

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

function mockApis(task: Task = fullTask) {
  vi.spyOn(tasksApi,    "getTask").mockResolvedValue(task);
  vi.spyOn(usersApi,    "listUsers").mockResolvedValue([assigneeUser]);
  vi.spyOn(agenciesApi, "getAgency").mockResolvedValue(agency);
}

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
      mockApis();
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("link", { name: /back/i })).toHaveAttribute("href", "/");
    });
  });

  describe("success state — required fields", () => {
    beforeEach(() => mockApis());

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
      expect(screen.getByText("In Progress", { selector: "span" })).toBeInTheDocument();
    });

    it("renders the priority badge", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText("High")).toBeInTheDocument();
    });

    it("renders the agency name", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(agency.name)).toBeInTheDocument();
    });

    it("renders the created_at date", async () => {
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/jan.*15.*2024|2024.*01.*15/i)).toBeInTheDocument();
    });
  });

  describe("success state — optional fields", () => {
    it("renders the description when present", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText("Users cannot log in with SSO enabled")).toBeInTheDocument();
    });

    it("shows no description when description is null", async () => {
      mockApis(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      expect(screen.getByText(/no description/i)).toBeInTheDocument();
    });

    it("renders the assignee name when present", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(assigneeUser.name)).toBeInTheDocument();
    });

    it("shows unassigned when assignee_id is null", async () => {
      mockApis(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      expect(screen.getByText(/unassigned/i)).toBeInTheDocument();
    });

    it("renders the formatted due date when present", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/feb.*1.*2024|2024.*02.*01/i)).toBeInTheDocument();
    });

    it("shows no due date text when due_date is null", async () => {
      mockApis(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      expect(screen.getByText(/no due date/i)).toBeInTheDocument();
    });

    it("renders the completed_at date when present", async () => {
      mockApis(completedTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/jan.*20.*2024|2024.*01.*20/i)).toBeInTheDocument();
    });

    it("shows not completed when completed_at is null", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByText(/not completed/i)).toBeInTheDocument();
    });
  });

  describe("assignee", () => {
    it("renders a select to change the assignee", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("combobox", { name: /assignee/i })).toBeInTheDocument();
    });

    it("populates the select with users", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("option", { name: assigneeUser.name })).toBeInTheDocument();
    });

    it("pre-selects the current assignee", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      const select = screen.getByRole("combobox", { name: /assignee/i }) as HTMLSelectElement;
      expect(select.value).toBe(assigneeId);
    });

    it("calls assignTask when a new assignee is selected", async () => {
      const spy = vi.spyOn(tasksApi, "assignTask").mockResolvedValue(undefined);
      const newUser: User = { id: "new-user-id", name: "Bob", email: "bob@acme.com", role: "member", agency_id: agencyId, created_at: "2024-01-01T00:00:00Z" };
      vi.spyOn(tasksApi, "getTask").mockResolvedValue(fullTask);
      vi.spyOn(usersApi, "listUsers").mockResolvedValue([assigneeUser, newUser]);
      vi.spyOn(agenciesApi, "getAgency").mockResolvedValue(agency);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      await userEvent.selectOptions(screen.getByRole("combobox", { name: /assignee/i }), "new-user-id");
      await waitFor(() => expect(spy).toHaveBeenCalledWith(taskId, { assignee_id: "new-user-id", assignee_agency_id: agencyId }));
    });
  });

  describe("due date", () => {
    it("renders a date input for the due date field", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByLabelText("due-date-input")).toBeInTheDocument();
    });

    it("pre-populates the date input with the current due date", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      // fullTask.due_date = "2024-02-01T00:00:00Z" → YYYY-MM-DD = "2024-02-01"
      const input = screen.getByLabelText("due-date-input") as HTMLInputElement;
      expect(input.value).toBe("2024-02-01");
    });

    it("renders an empty date input when due_date is null", async () => {
      mockApis(minimalTask);
      renderPage();
      await waitFor(() => screen.getByText("Minimal task"));
      const input = screen.getByLabelText("due-date-input") as HTMLInputElement;
      expect(input.value).toBe("");
    });

    it("calls updateDueDate when a new date is entered", async () => {
      const spy = vi.spyOn(tasksApi, "updateDueDate").mockResolvedValue(undefined);
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      const input = screen.getByLabelText("due-date-input");
      await userEvent.clear(input);
      await userEvent.type(input, "2024-03-15");
      input.blur();
      await waitFor(() =>
        expect(spy).toHaveBeenCalledWith(taskId, "2024-03-15")
      );
    });

    it("calls updateDueDate with null when the date is cleared", async () => {
      const spy = vi.spyOn(tasksApi, "updateDueDate").mockResolvedValue(undefined);
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      const input = screen.getByLabelText("due-date-input");
      await userEvent.clear(input);
      input.blur();
      await waitFor(() =>
        expect(spy).toHaveBeenCalledWith(taskId, null)
      );
    });
  });

  describe("status actions", () => {
    it("renders a status select", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("combobox", { name: /change status/i })).toBeInTheDocument();
    });

    it("has Complete and Set In Progress options", async () => {
      mockApis(fullTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("option", { name: /complete/i })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: /in progress/i })).toBeInTheDocument();
    });

    it("disables the Complete option when the task is already done", async () => {
      mockApis(completedTask);
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("option", { name: /complete/i })).toBeDisabled();
    });

    it("disables the Set In Progress option when the task is already in progress", async () => {
      mockApis(fullTask); // fullTask has status "in_progress"
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      expect(screen.getByRole("option", { name: /in progress/i })).toBeDisabled();
    });

    it("calls completeTask when Complete is selected", async () => {
      const spy = vi.spyOn(tasksApi, "completeTask").mockResolvedValue(undefined);
      mockApis({ ...fullTask, status: "todo" });
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      await userEvent.selectOptions(screen.getByRole("combobox", { name: /change status/i }), "complete");
      await waitFor(() => expect(spy).toHaveBeenCalledWith(taskId, expect.any(Object)));
    });

    it("calls setTaskInProgress when Set In Progress is selected", async () => {
      const spy = vi.spyOn(tasksApi, "setTaskInProgress").mockResolvedValue(undefined);
      mockApis({ ...fullTask, status: "todo" });
      renderPage();
      await waitFor(() => screen.getByText("Fix login bug"));
      await userEvent.selectOptions(screen.getByRole("combobox", { name: /change status/i }), "in_progress");
      await waitFor(() => expect(spy).toHaveBeenCalledWith(taskId, expect.any(Object)));
    });
  });
});
