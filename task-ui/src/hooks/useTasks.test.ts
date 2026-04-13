// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useTasks } from "./useTasks";
import * as tasksApi from "../api/tasks";
import type { Task } from "../types/api";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";

const tasks: Task[] = [
  {
    id: "c3d4e5f6-0000-0000-0000-000000000003",
    title: "Fix bug",
    description: null,
    status: "todo",
    priority: "high",
    agency_id: agencyId,
    assignee_id: null,
    created_at: "2024-01-01T00:00:00Z",
    due_date: null,
    completed_at: null,
  },
];

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("useTasks", () => {
  it("returns data from listTasks on success", async () => {
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

    const { result } = renderHook(() => useTasks(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(tasks);
  });

  it("calls listTasks with the given agencyId", async () => {
    const spy = vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

    const { result } = renderHook(() => useTasks(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(spy).toHaveBeenCalledWith(agencyId);
  });

  it("is in a loading state initially", () => {
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

    const { result } = renderHook(() => useTasks(agencyId), { wrapper });

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
  });

  it("sets isError when listTasks rejects", async () => {
    vi.spyOn(tasksApi, "listTasks").mockRejectedValue(new Error("network error"));

    const { result } = renderHook(() => useTasks(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
