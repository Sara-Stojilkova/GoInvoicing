import { describe, it, expect, vi, beforeEach } from "vitest";
import { loginApi, registerApi } from "./auth";

function makeFetchResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("loginApi", () => {
  it("POSTs to /api/auth/login with email and password", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      makeFetchResponse(200, {
        access_token: "tok",
        refresh_token: "ref",
        expires_in: 3600,
        token_type: "bearer",
        user: { email: "a@b.com", app_metadata: { agency_id: "ag-1", role: "admin" } },
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await loginApi({ email: "a@b.com", password: "pass" });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/auth/login");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({ email: "a@b.com", password: "pass" });
  });

  it("returns the parsed login response", async () => {
    const body = {
      access_token: "tok",
      refresh_token: "ref",
      expires_in: 3600,
      token_type: "bearer",
      user: { email: "a@b.com", app_metadata: { agency_id: "ag-1", role: "admin" } },
    };
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(200, body)));

    const result = await loginApi({ email: "a@b.com", password: "pass" });

    expect(result.access_token).toBe("tok");
    expect(result.user.app_metadata.agency_id).toBe("ag-1");
    expect(result.user.app_metadata.role).toBe("admin");
  });
});

describe("registerApi", () => {
  it("POSTs to /api/auth/register with the request body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      makeFetchResponse(201, { user_id: "u1", agency_id: "ag-1", role: "admin", activated: true })
    );
    vi.stubGlobal("fetch", fetchMock);

    await registerApi({ full_name: "Jane", email: "a@b.com", password: "pass", agency_name: "Acme" });

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/auth/register");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({
      full_name: "Jane",
      email: "a@b.com",
      password: "pass",
      agency_name: "Acme",
    });
  });

  it("returns the parsed register response", async () => {
    const body = { user_id: "u1", agency_id: "ag-1", role: "admin", activated: true };
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(201, body)));

    const result = await registerApi({ full_name: "Jane", email: "a@b.com", password: "pass", agency_name: "Acme" });

    expect(result.user_id).toBe("u1");
    expect(result.role).toBe("admin");
    expect(result.activated).toBe(true);
  });
});
