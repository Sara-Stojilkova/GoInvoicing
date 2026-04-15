import { describe, it, expect, vi, beforeEach } from "vitest";
import { listUsers, getUser, createUser } from "./users";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";
const user = {
  id: "b2c3d4e5-0000-0000-0000-000000000002",
  name: "Alice",
  email: "alice@acme.com",
  role: "admin" as const,
  agency_id: agencyId,
  created_at: "2024-01-01T00:00:00Z",
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

describe("listUsers", () => {
  it("calls GET /api/users with agency_id query param", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify([user]), { status: 200, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await listUsers(agencyId);

    expect(fetchMock.mock.calls[0][0]).toBe(`/api/users?agency_id=${agencyId}`);
  });

  it("returns an array of users", async () => {
    mockFetch(200, [user]);
    const result = await listUsers(agencyId);
    expect(result).toEqual([user]);
  });

  it("returns an empty array when there are no users", async () => {
    mockFetch(200, []);
    const result = await listUsers(agencyId);
    expect(result).toEqual([]);
  });

  it("throws ApiError with status 400 on invalid agency_id", async () => {
    mockFetch(400, { error: "invalid agency_id" });
    await expect(listUsers("bad")).rejects.toMatchObject({ status: 400 });
  });
});

describe("getUser", () => {
  it("calls GET /api/users/:id", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(user), { status: 200, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await getUser(user.id);

    expect(fetchMock.mock.calls[0][0]).toBe(`/api/users/${user.id}`);
  });

  it("returns the user", async () => {
    mockFetch(200, user);
    const result = await getUser(user.id);
    expect(result).toEqual(user);
  });

  it("throws ApiError with status 404 when not found", async () => {
    mockFetch(404, { error: "not found" });
    await expect(getUser("missing-id")).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError with status 400 on invalid id", async () => {
    mockFetch(400, { error: "invalid user id" });
    await expect(getUser("bad")).rejects.toMatchObject({ status: 400 });
  });
});

describe("createUser", () => {
  const payload = { name: "Alice", email: "alice@acme.com", role: "admin", agency_id: agencyId };

  it("calls POST /api/users", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(user), { status: 201, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await createUser(payload);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/users");
    expect(init.method).toBe("POST");
  });

  it("sends all fields in the request body", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(user), { status: 201, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await createUser(payload);

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(init.body as string)).toEqual(payload);
  });

  it("returns the created user", async () => {
    mockFetch(201, user);
    const result = await createUser(payload);
    expect(result).toEqual(user);
  });

  it("throws ApiError with status 400 when required fields are missing", async () => {
    mockFetch(400, { error: "name is required" });
    await expect(createUser({ ...payload, name: "" })).rejects.toMatchObject({ status: 400 });
  });

  it("throws ApiError with status 400 on invalid agency_id", async () => {
    mockFetch(400, { error: "invalid agency_id" });
    await expect(createUser({ ...payload, agency_id: "bad" })).rejects.toMatchObject({ status: 400 });
  });
});
