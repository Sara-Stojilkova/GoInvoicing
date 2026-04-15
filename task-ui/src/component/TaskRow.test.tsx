// @vitest-environment jsdom
import { describe, it, expect, vi , beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TaskRow } from "./TaskRow.tsx";
import type { Task } from "../types/api";

const task: Task = {
  id: "c3d4e5f6-0000-0000-0000-000000000003",
  title: "Fix login bug",
  description: null,
  status: "todo",
  priority: "high",
  agency_id: "a1b2c3d4-0000-0000-0000-000000000001",
  assignee_id: null,
  created_at: "2024-01-01T00:00:00Z",
  due_date: null,
  completed_at: null,
};

const mutate = vi.fn();
const useCompleteTaskMock = vi.fn();

vi.mock("../hooks/useTasks", () => ({
  useCompleteTask: () => useCompleteTaskMock(),
}));


beforeEach(() => {
  mutate.mockClear();
  useCompleteTaskMock.mockClear();
  useCompleteTaskMock.mockReturnValue({ mutate, isPending: false,});
});

function renderRow() {
  return render(
    <MemoryRouter>
      <table><tbody><tr><TaskRow task={task} /></tr></tbody></table>
    </MemoryRouter>
  );
}

describe("TaskRow", () => {
  it("renders the task title", () => {
    renderRow();
    expect(screen.getByText("Fix login bug")).toBeInTheDocument();
  });

  it("renders the task title as a link to the detail page", () => {
    renderRow();
    const link = screen.getByRole("link", { name: "Fix login bug" });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", `/tasks/${task.id}`);
  });

  it("renders the status badge", () => {
    renderRow();
    expect(screen.getByText("Todo").closest("span")?.className).toContain("badge-gray");
  });

  it("renders the priority", () => {
    renderRow();
    expect(screen.getByText("high")).toBeInTheDocument();
  });

  it("renders complete button", () => {
    renderRow();
    expect(screen.getByRole("button", { name: /complete/i })).toBeInTheDocument();
  });

  it("calls mutate when complete button is clicked", () => {
    useCompleteTaskMock.mockReturnValue({ mutate, isPending: false, });
    renderRow();
    fireEvent.click(screen.getByRole("button", { name: /complete/i }));
    expect(mutate).toHaveBeenCalledWith(task.id);
  });

  it("disables button while mutation is pending", () => {
    useCompleteTaskMock.mockReturnValue({ mutate, isPending: true, });
    renderRow();
    const button = screen.getByRole("button", { name: /complete|loading/i });
    expect(button).toBeDisabled();
  });

  it("shows spinner when completing task", () => {
    useCompleteTaskMock.mockReturnValue({ mutate, isPending: true, });
    renderRow();
    const button = screen.getByRole("button", { name: /complete|loading/i });
    expect(button).toBeDisabled();
    expect(button.querySelector("span")).toBeInTheDocument();
  });
});