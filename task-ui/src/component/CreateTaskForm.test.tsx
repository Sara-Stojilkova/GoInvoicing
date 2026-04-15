// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor} from "@testing-library/react";
import { CreateTaskForm } from "./CreateTaskForm";

const mutate = vi.fn((_, options) => { options?.onSuccess?.(); });
const mockUseCreateTask = vi.fn();

vi.mock("../hooks/useTasks", () => ({
  useCreateTask: () => mockUseCreateTask(),
}));

beforeEach(() => {
  mutate.mockClear();
  mockUseCreateTask.mockClear();
  mockUseCreateTask.mockReturnValue({
    mutate,
    isPending: false,
    isError: false,
    error: null,
  });
});

describe("CreateTaskForm", () => {
  it("renders form inputs", () => {
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByLabelText(/title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/priority/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/assignee id/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/due date/i)).toBeInTheDocument();
  });

  it("does not submit when required fields are empty", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).not.toHaveBeenCalled();
  });

  it("submits correct data", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" } });
    fireEvent.change(screen.getByLabelText(/priority/i), { target: { value: "high" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).toHaveBeenCalledWith({
      title: "New task",
      priority: "high",
      agency_id: "a1",
      },
      expect.any(Object)
    );
  });

  it("disables submit while pending", () => {
    mockUseCreateTask.mockReturnValue({ mutate, isPending: true, isError: false, error: null });
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByRole("button", { name: /create/i })).toBeDisabled();
  });

  it("shows error message on failure", () => {
    mockUseCreateTask.mockReturnValue({
      mutate,
      isPending: false,
      isError: true,
      error: new Error("Failed to create task"),
    });
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByText(/failed/i)).toBeInTheDocument();
  });

  it("resets form after submit success", async () => {
    render(<CreateTaskForm agencyId="a1" />);
    const title = screen.getByLabelText(/title/i);
    const priority = screen.getByLabelText(/priority/i);
    fireEvent.change(title, { target: { value: "New task" } });
    fireEvent.change(priority, { target: { value: "high" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    await waitFor(() => { 
      expect(title).toHaveValue("");
      expect(priority).toHaveValue("medium"); 
    });
  });

  it("shows validation error when title is empty", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(screen.getByText(/title is required/i)).toBeInTheDocument();
    expect(mutate).not.toHaveBeenCalled();
  });

  it("clears validation error when input becomes valid", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(screen.getByText(/title is required/i)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" }, });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(screen.queryByText(/title is required/i)).not.toBeInTheDocument();
  });

  it("includes optional fields when provided", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" }, });
    fireEvent.change(screen.getByLabelText(/description/i), { target: { value: "test desc" }, });
    fireEvent.change(screen.getByLabelText(/assignee id/i), { target: { value: "user-1" }, });
    fireEvent.change(screen.getByLabelText(/due date/i), { target: { value: "2026-01-01" }, });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).toHaveBeenCalledWith(
    {
      title: "New task",
      priority: "medium",
      agency_id: "a1",
      description: "test desc",
      assignee_id: "user-1",
      due_date: "2026-01-01",
    },
      expect.any(Object)
    );
  });
});