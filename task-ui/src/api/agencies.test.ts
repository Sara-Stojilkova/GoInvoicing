import { describe, it, expect, vi, beforeEach } from "vitest";
import { listAgencies, getAgency, createAgency } from "./agencies";

const agency = { id: "a1b2c3d4-0000-0000-0000-000000000001", name: "Acme", created_at: "2024-01-01T00:00:00Z" };

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

describe("listAgencies", () => {
  it("calls GET /api/agencies", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify([agency]), { status: 200, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await listAgencies();

    expect(fetchMock.mock.calls[0][0]).toBe("/api/agencies");
    expect(fetchMock.mock.calls[0][1]?.method ?? "GET").toBe("GET");
  });

  it("returns an array of agencies", async () => {
    mockFetch(200, [agency]);
    const result = await listAgencies();
    expect(result).toEqual([agency]);
  });

  it("returns an empty array when there are no agencies", async () => {
    mockFetch(200, []);
    const result = await listAgencies();
    expect(result).toEqual([]);
  });

  it("throws ApiError on 500", async () => {
    mockFetch(500, { error: "internal server error" });
    await expect(listAgencies()).rejects.toMatchObject({ status: 500 });
  });
});

describe("getAgency", () => {
  it("calls GET /api/agencies/:id", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(agency), { status: 200, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await getAgency(agency.id);

    expect(fetchMock.mock.calls[0][0]).toBe(`/api/agencies/${agency.id}`);
  });

  it("returns the agency", async () => {
    mockFetch(200, agency);
    const result = await getAgency(agency.id);
    expect(result).toEqual(agency);
  });

  it("throws ApiError with status 404 when not found", async () => {
    mockFetch(404, { error: "not found" });
    await expect(getAgency("missing-id")).rejects.toMatchObject({ status: 404 });
  });

  it("throws ApiError with status 400 on invalid id", async () => {
    mockFetch(400, { error: "invalid agency id" });
    await expect(getAgency("bad")).rejects.toMatchObject({ status: 400 });
  });
});

describe("createAgency", () => {
  it("calls POST /api/agencies", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(agency), { status: 201, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await createAgency("Acme");

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/agencies");
    expect(init.method).toBe("POST");
  });

  it("sends the name in the request body", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      new Response(JSON.stringify(agency), { status: 201, headers: { "Content-Type": "application/json" } })
    );
    vi.stubGlobal("fetch", fetchMock);

    await createAgency("Acme");

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(init.body as string)).toEqual({ name: "Acme" });
  });

  it("returns the created agency", async () => {
    mockFetch(201, agency);
    const result = await createAgency("Acme");
    expect(result).toEqual(agency);
  });

  it("throws ApiError with status 400 when name is missing", async () => {
    mockFetch(400, { error: "name is required" });
    await expect(createAgency("")).rejects.toMatchObject({ status: 400 });
  });
});
