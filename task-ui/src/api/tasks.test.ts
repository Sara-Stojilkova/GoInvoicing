import { describe, it, expect, vi, beforeEach } from "vitest";
import { listTasks, getTask, createTask, assignTask, completeTask, setTaskInProgress } from "./tasks";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";
const taskId = "c3d4e5f6-0000-0000-0000-000000000003";
const task = {
  id: taskId,
  title: "Fix bug",
  description: null,
  status: "todo" as const,
  priority: "high" as const,
  agency_id: agencyId,
  assignee_id: null,
  created_at: "2024-01-01T00:00:00Z",
  due_date: null,
  completed_at: null,
};

function mockFetch(status: number, body: unknown) {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
      })
    )
  );
}

beforeEach(() => vi.restoreAllMocks());

describe("listTasks", () => {
  it("calls GET /api/tasks with agency_id query param", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify([task]), { status: 200, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await listTasks(agencyId);

    expect(fetchMock.mock.calls[0][0]).toBe(`/api/tasks?agency_id=${agencyId}`);
  });

  it("returns an array of tasks", async () => {
    mockFetch(200, [task]);
    const result = await listTasks(agencyId);
    expect(result).toEqual([task]);
  });

  it("returns an empty array when there are no tasks", async () => {
    mockFetch(200, []);
    const result = await listTasks(agencyId);
    expect(result).toEqual([]);
  });

  it("throws ApiError with status 400 on missing agency_id", async () => {
    mockFetch(400, { error: "agency_id query param is required and must be a valid UUID" });
    await expect(listTasks("")).rejects.toMatchObject({ status: 400 });
  });
});

describe("getTask", () => {
  it("calls GET /api/tasks/:id with agency_id query param", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(task), { status: 200, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await getTask(taskId, agencyId);

    expect(fetchMock.mock.calls[0][0]).toBe(`/api/tasks/${taskId}?agency_id=${agencyId}`);
  });

  it("returns the task", async () => {
    mockFetch(200, task);
    const result = await getTask(taskId, agencyId);
    expect(result).toEqual(task);
  });

  it("throws ApiError with status 404 when not found", async () => {
    mockFetch(404, { error: "not found" });
    await expect(getTask("missing-id", agencyId)).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError with status 403 on agency mismatch", async () => {
    mockFetch(403, { error: "forbidden" });
    await expect(getTask(taskId, "other-agency")).rejects.toMatchObject({ status: 403 });
  });
});

describe("createTask", () => {
  const payload = { title: "Fix bug", priority: "high", agency_id: agencyId };

  it("calls POST /api/tasks", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(task), { status: 201, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await createTask(payload);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/tasks");
    expect(init.method).toBe("POST");
  });

  it("sends required fields in the request body", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(task), { status: 201, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await createTask(payload);

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    const body = JSON.parse(init.body as string);
    expect(body.title).toBe(payload.title);
    expect(body.priority).toBe(payload.priority);
    expect(body.agency_id).toBe(payload.agency_id);
  });

  it("returns the created task", async () => {
    mockFetch(201, task);
    const result = await createTask(payload);
    expect(result).toEqual(task);
  });

  it("throws ApiError with status 400 when title is missing", async () => {
    mockFetch(400, { error: "title is required" });
    await expect(createTask({ ...payload, title: "" })).rejects.toMatchObject({ status: 400 });
  });

  it("sends optional fields when provided", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(task), { status: 201, headers: { "Content-Type": "application/json" }, })
    );
    vi.stubGlobal("fetch", fetchMock);
    await createTask({ ...payload, description: "test desc", due_date: "2026-01-01", });
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(body.description).toBe("test desc");
    expect(body.due_date).toBe("2026-01-01T00:00:00.000Z");
  });
});

describe("assignTask", () => {
  const assigneeId = "d4e5f6a7-0000-0000-0000-000000000004";
  const payload = { assignee_id: assigneeId, assignee_agency_id: agencyId };

  it("calls POST /api/tasks/:id/assign", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(null, { status: 204 })
    );
    vi.stubGlobal("fetch", fetchMock);

    await assignTask(taskId, payload);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`/api/tasks/${taskId}/assign`);
    expect(init.method).toBe("POST");
  });

  it("sends assignee_id and assignee_agency_id in the request body", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(null, { status: 204 })
    );
    vi.stubGlobal("fetch", fetchMock);

    await assignTask(taskId, payload);

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(init.body as string)).toEqual(payload);
  });

  it("returns void on 204", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(() => new Response(null, { status: 204 })));
    const result = await assignTask(taskId, payload);
    expect(result).toBeUndefined();
  });

  it("throws ApiError with status 403 on agency mismatch", async () => {
    mockFetch(403, { error: "forbidden" });
    await expect(assignTask(taskId, { ...payload, assignee_agency_id: "other" })).rejects.toMatchObject({ status: 403 });
  });

  it("throws ApiError with status 404 when task not found", async () => {
    mockFetch(404, { error: "not found" });
    await expect(assignTask("missing-id", payload)).rejects.toMatchObject({ status: 404 });
  });
});

describe("completeTask", () => {
  it("calls POST /api/tasks/:id/complete", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(null, { status: 204 })
    );
    vi.stubGlobal("fetch", fetchMock);

    await completeTask(taskId);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`/api/tasks/${taskId}/complete`);
    expect(init.method).toBe("POST");
  });

  it("returns void on 204", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(() => new Response(null, { status: 204 })));
    const result = await completeTask(taskId);
    expect(result).toBeUndefined();
  });

  it("throws ApiError with status 404 when task not found", async () => {
    mockFetch(404, { error: "not found" });
    await expect(completeTask("missing-id")).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError with status 409 when task is already completed", async () => {
    mockFetch(409, { error: "task already completed" });
    await expect(completeTask(taskId)).rejects.toMatchObject({ status: 409 });
  });
});

describe("setTaskInProgress", () => {
  it("calls POST /api/tasks/:id/set-in-progress", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(null, { status: 204 })
    );
    vi.stubGlobal("fetch", fetchMock);

    await setTaskInProgress(taskId);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`/api/tasks/${taskId}/set-in-progress`);
    expect(init.method).toBe("POST");
  });

  it("returns void on 204", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(() => new Response(null, { status: 204 })));
    const result = await setTaskInProgress(taskId);
    expect(result).toBeUndefined();
  });

  it("throws ApiError with status 404 when task not found", async () => {
    mockFetch(404, { error: "not found" });
    await expect(setTaskInProgress("missing-id")).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError with status 409 when task is already in progress", async () => {
    mockFetch(409, { error: "task already in progress" });
    await expect(setTaskInProgress(taskId)).rejects.toMatchObject({ status: 409 });
  });
});
