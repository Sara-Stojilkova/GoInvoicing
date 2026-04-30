import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { CreateTaskForm } from "./CreateTaskForm";
import type { User } from "../types/api";

const mutate = vi.fn((_, options) => { options?.onSuccess?.(); });
const mockUseCreateTask = vi.fn();
const mockUseUsers = vi.fn();

vi.mock("../hooks/useTasks", () => ({
  useCreateTask: () => mockUseCreateTask(),
}));

vi.mock("../hooks/useUsers", () => ({
  useUsers: () => mockUseUsers(),
}));

const users: User[] = [
  { id: "user-uuid-1", name: "Alice", email: "alice@acme.com", role: "admin", agency_id: "a1", created_at: "2024-01-01T00:00:00Z" },
  { id: "user-uuid-2", name: "Bob",   email: "bob@acme.com",   role: "member", agency_id: "a1", created_at: "2024-01-02T00:00:00Z" },
];

beforeEach(() => {
  mutate.mockClear();
  mockUseCreateTask.mockClear();
  mockUseCreateTask.mockReturnValue({ mutate, isPending: false, isError: false, error: null });
  mockUseUsers.mockClear();
  mockUseUsers.mockReturnValue({ data: users, isLoading: false, isError: false });
});

describe("CreateTaskForm", () => {
  it("renders form inputs", () => {
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByLabelText(/title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/priority/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/assignee/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/due date/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create/i })).toBeInTheDocument();
  });

  it("renders the assignee field as a select, not a text input", () => {
    render(<CreateTaskForm agencyId="a1" />);
    const assignee = screen.getByLabelText(/assignee/i);
    expect(assignee.tagName).toBe("SELECT");
  });

  it("populates the assignee select with users from the hook", () => {
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByRole("option", { name: "Alice" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Bob" })).toBeInTheDocument();
  });

  it("includes an unassigned option as the default", () => {
    render(<CreateTaskForm agencyId="a1" />);
    const unassigned = screen.getByRole("option", { name: /unassigned/i });
    expect(unassigned).toBeInTheDocument();
    expect((screen.getByLabelText(/assignee/i) as HTMLSelectElement).value).toBe("");
  });

  it("disables the assignee select while users are loading", () => {
    mockUseUsers.mockReturnValue({ data: undefined, isLoading: true, isError: false });
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByLabelText(/assignee/i)).toBeDisabled();
  });

  it("does not submit when required fields are empty", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).not.toHaveBeenCalled();
  });

  it("submits correct data without an assignee", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" } });
    fireEvent.change(screen.getByLabelText(/priority/i), { target: { value: "high" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).toHaveBeenCalledWith(
      { title: "New task", priority: "high", agency_id: "a1" },
      expect.any(Object)
    );
  });

  it("omits assignee_id from payload when unassigned is selected", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" } });
    fireEvent.change(screen.getByLabelText(/assignee/i), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    const payload = mutate.mock.calls[0][0];
    expect(payload.assignee_id).toBeUndefined();
  });

  it("includes assignee_id in payload when a user is selected", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" } });
    fireEvent.change(screen.getByLabelText(/assignee/i), { target: { value: "user-uuid-1" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).toHaveBeenCalledWith(
      expect.objectContaining({ assignee_id: "user-uuid-1" }),
      expect.any(Object)
    );
  });

  it("disables submit while pending", () => {
    mockUseCreateTask.mockReturnValue({ mutate, isPending: true, isError: false, error: null });
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByRole("button", { name: /create/i })).toBeDisabled();
  });

  it("shows error message on failure", () => {
    mockUseCreateTask.mockReturnValue({ mutate, isPending: false, isError: true, error: new Error("Failed to create task") });
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByText(/failed/i)).toBeInTheDocument();
  });

  it("resets form after submit success", async () => {
    render(<CreateTaskForm agencyId="a1" />);
    const title = screen.getByLabelText(/title/i);
    const priority = screen.getByLabelText(/priority/i);
    const assignee = screen.getByLabelText(/assignee/i);
    fireEvent.change(title, { target: { value: "New task" } });
    fireEvent.change(priority, { target: { value: "high" } });
    fireEvent.change(assignee, { target: { value: "user-uuid-1" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    await waitFor(() => {
      expect(title).toHaveValue("");
      expect(priority).toHaveValue("medium");
      expect((assignee as HTMLSelectElement).value).toBe("");
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
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(screen.queryByText(/title is required/i)).not.toBeInTheDocument();
  });

  it("includes optional fields when provided", () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "New task" } });
    fireEvent.change(screen.getByLabelText(/description/i), { target: { value: "test desc" } });
    fireEvent.change(screen.getByLabelText(/assignee/i), { target: { value: "user-uuid-2" } });
    fireEvent.change(screen.getByLabelText(/due date/i), { target: { value: "2026-01-01" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    expect(mutate).toHaveBeenCalledWith(
      {
        title: "New task",
        priority: "medium",
        agency_id: "a1",
        description: "test desc",
        assignee_id: "user-uuid-2",
        due_date: "2026-01-01",
      },
      expect.any(Object)
    );
  });

  it("renders a tags input", () => {
    render(<CreateTaskForm agencyId="a1" />);
    expect(screen.getByPlaceholderText(/add a tag/i)).toBeInTheDocument();
  });

  it("submits with tags when tags are added", async () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "My task" } });

    const tagInput = screen.getByPlaceholderText(/add a tag/i);
    fireEvent.change(tagInput, { target: { value: "bug" } });
    fireEvent.keyDown(tagInput, { key: "Enter" });

    fireEvent.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() => {
      expect(mutate).toHaveBeenCalledWith(
        expect.objectContaining({ tags: ["bug"] }),
        expect.anything()
      );
    });
  });

  it("submits without tags when none added", async () => {
    render(<CreateTaskForm agencyId="a1" />);
    fireEvent.change(screen.getByLabelText(/title/i), { target: { value: "My task" } });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));
    await waitFor(() => {
      const call = mutate.mock.calls[mutate.mock.calls.length - 1][0];
      expect(call.tags).toBeUndefined();
    });
  });
});
