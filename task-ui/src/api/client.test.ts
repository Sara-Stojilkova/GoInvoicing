import { describe, it, expect, vi, beforeEach } from "vitest";
import { request } from "./client";
import { ApiError } from "./error";

function makeFetchResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

beforeEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

describe("request", () => {
  it("returns parsed JSON on a 200 response", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(200, { id: "1", name: "Acme" })));

    const result = await request<{ id: string; name: string }>("/agencies/1");

    expect(result).toEqual({ id: "1", name: "Acme" });
  });

  it("throws ApiError with the response status on a 404", async () => {
    vi.stubGlobal("fetch", vi.fn().mockImplementation(() => makeFetchResponse(404, { error: "not found" })));

    await expect(request("/agencies/999")).rejects.toBeInstanceOf(ApiError);
    await expect(request("/agencies/999")).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError with the response status on a 500", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(500, { error: "internal server error" })));

    await expect(request("/tasks")).rejects.toMatchObject({ status: 500 });
  });

  it("uses the error message from the response body", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(403, { error: "forbidden" })));

    await expect(request("/tasks")).rejects.toMatchObject({ message: "forbidden" });
  });

  it("sets Content-Type: application/json on every request", async () => {
    const fetchMock = vi.fn().mockResolvedValue(makeFetchResponse(200, {}));
    vi.stubGlobal("fetch", fetchMock);

    await request("/tasks");

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect((init.headers as Record<string, string>)["Content-Type"]).toBe("application/json");
  });

  it("forwards the HTTP method and body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(makeFetchResponse(200, { id: "abc" }));
    vi.stubGlobal("fetch", fetchMock);

    await request("/tasks", { method: "POST", body: JSON.stringify({ title: "Fix bug" }) });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/tasks");
    expect(init.method).toBe("POST");
    expect(init.body).toBe(JSON.stringify({ title: "Fix bug" }));
  });

  it("passes the URL through to fetch unchanged", async () => {
    const fetchMock = vi.fn().mockResolvedValue(makeFetchResponse(200, []));
    vi.stubGlobal("fetch", fetchMock);

    await request("/users?agency_id=1");

    expect(fetchMock.mock.calls[0][0]).toBe("/users?agency_id=1");
  });

  it("attaches Authorization header when auth_token is in localStorage", async () => {
    const fetchMock = vi.fn().mockResolvedValue(makeFetchResponse(200, {}));
    vi.stubGlobal("fetch", fetchMock);
    localStorage.setItem("auth_token", "my-token");

    await request("/tasks");

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect((init.headers as Record<string, string>)["Authorization"]).toBe("Bearer my-token");
  });

  it("does not attach Authorization header when no token in localStorage", async () => {
    const fetchMock = vi.fn().mockResolvedValue(makeFetchResponse(200, {}));
    vi.stubGlobal("fetch", fetchMock);

    await request("/tasks");

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect((init.headers as Record<string, string>)["Authorization"]).toBeUndefined();
  });

  it("dispatches auth:logout event on 401", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(401, { error: "unauthorized" })));
    const dispatchSpy = vi.spyOn(window, "dispatchEvent");

    await expect(request("/tasks")).rejects.toMatchObject({ status: 401 });

    expect(dispatchSpy).toHaveBeenCalledWith(expect.objectContaining({ type: "auth:logout" }));
  });

  it("does not dispatch auth:logout on non-401 errors", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(403, { error: "forbidden" })));
    const dispatchSpy = vi.spyOn(window, "dispatchEvent");

    await expect(request("/tasks")).rejects.toMatchObject({ status: 403 });

    expect(dispatchSpy).not.toHaveBeenCalledWith(expect.objectContaining({ type: "auth:logout" }));
  });
});
