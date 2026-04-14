// @vitest-environment jsdom
import { describe, it, expect, vi , beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
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

describe("TaskRow", () => {
  it("renders the task title", () => {
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    expect(screen.getByText("Fix login bug")).toBeInTheDocument();
  });

  it("renders the status badge", () => {
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    expect(screen.getByText("Todo").closest("span")?.className).toContain("badge-gray");
  });

  it("renders the priority", () => {
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    expect(screen.getByText("high")).toBeInTheDocument();
  });

  it("renders complete button", () => {
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    expect(screen.getByRole("button", { name: /complete/i })).toBeInTheDocument();
  });

  it("calls mutate when complete button is clicked", () => {
    useCompleteTaskMock.mockReturnValue({ mutate, isPending: false, });
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    fireEvent.click(screen.getByRole("button", { name: /complete/i }));
    expect(mutate).toHaveBeenCalledWith(task.id);
  });

  it("disables button while mutation is pending", () => {
    useCompleteTaskMock.mockReturnValue({ mutate, isPending: true, });
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    const button = screen.getByRole("button", { name: /complete/i });
    expect(button).toBeDisabled();
  });

  it("shows spinner when completing task", () => {
    useCompleteTaskMock.mockReturnValue({ mutate, isPending: true, });
    render(<table><tbody><tr><TaskRow task={task} /></tr></tbody></table>);
    const button = screen.getByRole("button", { name: /complete/i });
    expect(button).toBeDisabled();
    expect(button.querySelector("span")).toBeInTheDocument();
  });
});