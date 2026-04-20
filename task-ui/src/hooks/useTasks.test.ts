import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useTasks, useCreateTask, useCompleteTask, useTask, useAssignTask, useSetInProgress, useUpdateDueDate, useUpdateDescription } from "./useTasks";
import { createTestQueryClient } from "../test/testQueryClient";
import { createWrapper } from "../test/wrapper";
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
  const queryClient = createTestQueryClient({ gcTime: Infinity });
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
  return { queryClient, Wrapper };
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("useTasks", () => {
  it("returns data from listTasks on success", async () => {
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

    const { result } = renderHook(() => useTasks(agencyId), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toEqual(tasks);
  });

  it("calls listTasks with the given agencyId", async () => {
    const spy = vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

    const { result } = renderHook(() => useTasks(agencyId), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(spy).toHaveBeenCalledWith(agencyId);
  });

  it("is in a loading state initially", () => {
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);

    const { result } = renderHook(() => useTasks(agencyId), { wrapper: createWrapper() });

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
  });

  it("sets isError when listTasks rejects", async () => {
    vi.spyOn(tasksApi, "listTasks").mockRejectedValue(new Error("network error"));

    const { result } = renderHook(() => useTasks(agencyId), { wrapper: createWrapper() });

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
    expect(spy).toHaveBeenCalledWith(payload, expect.any(Object));
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

describe("useCompleteTask", () => {
  it("calls completeTask with the task id", async () => {
    const spy = vi.spyOn(tasksApi, "completeTask").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useCompleteTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, expect.any(Object));
  });

  it("invalidates the tasks list cache for the agency on success", async () => {
    vi.spyOn(tasksApi, "completeTask").mockResolvedValue(undefined);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", agencyId],
      queryFn: () => tasksApi.listTasks(agencyId),
    });
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(false);

    const { result } = renderHook(() => useCompleteTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(true);
  });

  it("does not invalidate a different agency's task cache", async () => {
    const otherAgencyId = "ffffffff-0000-0000-0000-000000000099";
    vi.spyOn(tasksApi, "completeTask").mockResolvedValue(undefined);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", otherAgencyId],
      queryFn: () => tasksApi.listTasks(otherAgencyId),
    });

    const { result } = renderHook(() => useCompleteTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", otherAgencyId])?.isInvalidated).toBe(false);
  });

  it("exposes isError when completeTask rejects", async () => {
    vi.spyOn(tasksApi, "completeTask").mockRejectedValue(new Error("already completed"));
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useCompleteTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useTask", () => {
  it("returns the task when given a valid id", async () => {
    vi.spyOn(tasksApi, "getTask").mockResolvedValue(tasks[0]);

    const { result } = renderHook(() => useTask(tasks[0].id, agencyId), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(tasks[0]);
  });

  it("calls getTask with the task id and agency id", async () => {
    const spy = vi.spyOn(tasksApi, "getTask").mockResolvedValue(tasks[0]);

    const { result } = renderHook(() => useTask(tasks[0].id, agencyId), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, agencyId);
  });

  it("does not fetch when taskId is null", () => {
    const spy = vi.spyOn(tasksApi, "getTask").mockResolvedValue(tasks[0]);

    const { result } = renderHook(() => useTask(null, agencyId), { wrapper: createWrapper() });

    expect(spy).not.toHaveBeenCalled();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeUndefined();
  });

  it("fetches when taskId changes from null to a valid id", async () => {
    const spy = vi.spyOn(tasksApi, "getTask").mockResolvedValue(tasks[0]);
    const { Wrapper } = makeWrapper();

    const { result, rerender } = renderHook(
      ({ id }: { id: string | null }) => useTask(id, agencyId),
      { wrapper: Wrapper, initialProps: { id: null as string | null } }
    );

    expect(spy).not.toHaveBeenCalled();

    rerender({ id: tasks[0].id });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, agencyId);
  });

  it("sets isError when getTask rejects", async () => {
    vi.spyOn(tasksApi, "getTask").mockRejectedValue(new Error("not found"));

    const { result } = renderHook(() => useTask(tasks[0].id, agencyId), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useAssignTask", () => {
  const assigneeId = "b2c3d4e5-0000-0000-0000-000000000002";
  const payload = { taskId: tasks[0].id, assigneeId, assigneeAgencyId: agencyId };

  it("calls assignTask with the task and assignee ids", async () => {
    const spy = vi.spyOn(tasksApi, "assignTask").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useAssignTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, { assignee_id: assigneeId, assignee_agency_id: agencyId });
  });

  it("invalidates the task cache for the agency on success", async () => {
    vi.spyOn(tasksApi, "assignTask").mockResolvedValue(undefined);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", agencyId],
      queryFn: () => tasksApi.listTasks(agencyId),
    });
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(false);

    const { result } = renderHook(() => useAssignTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(true);
  });

  it("exposes isError when assignTask rejects", async () => {
    vi.spyOn(tasksApi, "assignTask").mockRejectedValue(new Error("not found"));
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useAssignTask(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useUpdateDescription", () => {
  const payload = { taskId: tasks[0].id, description: "Fix the login flow" };

  it("calls updateDescription with the task id and description", async () => {
    const spy = vi.spyOn(tasksApi, "updateDescription").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useUpdateDescription(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, "Fix the login flow");
  });

  it("calls updateDescription with null when description is cleared", async () => {
    const spy = vi.spyOn(tasksApi, "updateDescription").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useUpdateDescription(agencyId), { wrapper: Wrapper });
    result.current.mutate({ taskId: tasks[0].id, description: null });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, null);
  });

  it("invalidates the task cache for the agency on success", async () => {
    vi.spyOn(tasksApi, "updateDescription").mockResolvedValue(undefined);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", agencyId],
      queryFn: () => tasksApi.listTasks(agencyId),
    });
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(false);

    const { result } = renderHook(() => useUpdateDescription(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(true);
  });

  it("exposes isError when updateDescription rejects", async () => {
    vi.spyOn(tasksApi, "updateDescription").mockRejectedValue(new Error("not found"));
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useUpdateDescription(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useUpdateDueDate", () => {
  const payload = { taskId: tasks[0].id, dueDate: "2024-03-15" };

  it("calls updateDueDate with the task id and date", async () => {
    const spy = vi.spyOn(tasksApi, "updateDueDate").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useUpdateDueDate(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, "2024-03-15");
  });

  it("calls updateDueDate with null when the date is cleared", async () => {
    const spy = vi.spyOn(tasksApi, "updateDueDate").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useUpdateDueDate(agencyId), { wrapper: Wrapper });
    result.current.mutate({ taskId: tasks[0].id, dueDate: null });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, null);
  });

  it("invalidates the task cache for the agency on success", async () => {
    vi.spyOn(tasksApi, "updateDueDate").mockResolvedValue(undefined);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", agencyId],
      queryFn: () => tasksApi.listTasks(agencyId),
    });
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(false);

    const { result } = renderHook(() => useUpdateDueDate(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(true);
  });

  it("exposes isError when updateDueDate rejects", async () => {
    vi.spyOn(tasksApi, "updateDueDate").mockRejectedValue(new Error("not found"));
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useUpdateDueDate(agencyId), { wrapper: Wrapper });
    result.current.mutate(payload);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useSetInProgress", () => {
  it("calls setTaskInProgress with the task id", async () => {
    const spy = vi.spyOn(tasksApi, "setTaskInProgress").mockResolvedValue(undefined);
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useSetInProgress(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith(tasks[0].id, expect.any(Object));
  });

  it("invalidates the task cache for the agency on success", async () => {
    vi.spyOn(tasksApi, "setTaskInProgress").mockResolvedValue(undefined);
    vi.spyOn(tasksApi, "listTasks").mockResolvedValue(tasks);
    const { queryClient, Wrapper } = makeWrapper();

    await queryClient.prefetchQuery({
      queryKey: ["tasks", agencyId],
      queryFn: () => tasksApi.listTasks(agencyId),
    });
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(false);

    const { result } = renderHook(() => useSetInProgress(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryClient.getQueryState(["tasks", agencyId])?.isInvalidated).toBe(true);
  });

  it("exposes isError when setTaskInProgress rejects", async () => {
    vi.spyOn(tasksApi, "setTaskInProgress").mockRejectedValue(new Error("conflict"));
    const { Wrapper } = makeWrapper();

    const { result } = renderHook(() => useSetInProgress(agencyId), { wrapper: Wrapper });
    result.current.mutate(tasks[0].id);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
