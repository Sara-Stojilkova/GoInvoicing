import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
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
});