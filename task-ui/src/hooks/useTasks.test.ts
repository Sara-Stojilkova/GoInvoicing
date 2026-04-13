// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useTasks, useCreateTask } from "./useTasks";
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

const newTask: Task = {
  id: "d4e5f6a7-0000-0000-0000-000000000004",
  title: "New task",
  description: null,
  status: "todo",
  priority: "medium",
  agency_id: agencyId,
  assignee_id: null,
  created_at: "2024-01-02T00:00:00Z",
  due_date: null,
  completed_at: null,
};

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
  return { queryClient, Wrapper };
}

// kept for useTasks tests which don't need access to the queryClient
function wrapper({ children }: { children: React.ReactNode }) {
  return makeWrapper().Wrapper({ children });
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

describe("useCreateTask", () => {
  const payload = { title: "New task", priority: "medium", agency_id: agencyId };

  it("calls createTask with the supplied payload", async () => {
    const spy = vi.spyOn(tasksApi, "createTask").mockResolvedValue(newTask);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useCreateTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(payload);
  });

  it("invalidates the tasks list cache for the agency on success", async () => {
    vi.spyOn(tasksApi, "createTask").mockResolvedValue(newTask);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue([]);
    const { queryClient, Wrapper } = makeWrapper();

    // seed the cache so there is something to invalidate
    await queryClient.prefetchQuery({
      queryKey: ["tasks", agencyId],
      queryFn: () => tasksApi.listTasks(agencyId),
    });
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(false);

    const { result } = renderHook(() => useCreateTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(true);
  });

  it("does not invalidate a different agency's task cache", async () => {
    const otherAgencyId = "ffffffff-0000-0000-0000-000000000099";
    vi.spyOn(tasksApi, "createTask").mockResolvedValue(newTask);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue([]);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", otherAgencyId],
      queryFn: () => tasksApi.listTasks(otherAgencyId),
    });

    const { result } = renderHook(() => useCreateTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", otherAgencyId])?.isInvalidated).toBe(false);
  });

  it("exposes isError when createTask rejects", async () => {
    vi.spyOn(tasksApi, "createTask").mockRejectedValue(new Error("server error"));
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useCreateTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
