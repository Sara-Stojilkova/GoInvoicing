import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { AuthProvider, useAuth } from "./AuthContext";

function makeFetchResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

const loginBody = {
  access_token: "tok123",
  refresh_token: "ref456",
  expires_in: 3600,
  token_type: "bearer",
  user: {
    email: "a@b.com",
    app_metadata: { agency_id: "ag-uuid", role: "admin" },
  },
};

function TestConsumer() {
  const { token, agencyId, role, userEmail, login, logout } = useAuth();
  return (
    <div>
      <span data-testid="token">{token ?? "none"}</span>
      <span data-testid="agencyId">{agencyId ?? "none"}</span>
      <span data-testid="role">{role ?? "none"}</span>
      <span data-testid="email">{userEmail ?? "none"}</span>
      <button onClick={() => login("a@b.com", "pass")}>login</button>
      <button onClick={logout}>logout</button>
    </div>
  );
}

beforeEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

describe("AuthContext", () => {
  it("starts with null state when localStorage is empty", () => {
    render(<AuthProvider><TestConsumer /></AuthProvider>);
    expect(screen.getByTestId("token").textContent).toBe("none");
    expect(screen.getByTestId("agencyId").textContent).toBe("none");
  });

  it("restores session from localStorage on mount", () => {
    localStorage.setItem("auth_token", "saved-tok");
    localStorage.setItem("auth_agency_id", "saved-ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "a@b.com");

    render(<AuthProvider><TestConsumer /></AuthProvider>);

    expect(screen.getByTestId("token").textContent).toBe("saved-tok");
    expect(screen.getByTestId("agencyId").textContent).toBe("saved-ag");
    expect(screen.getByTestId("role").textContent).toBe("admin");
    expect(screen.getByTestId("email").textContent).toBe("a@b.com");
  });

  it("login saves token, agencyId, role, and email to state and localStorage", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(200, loginBody)));
    render(<AuthProvider><TestConsumer /></AuthProvider>);

    await act(async () => {
      screen.getByRole("button", { name: "login" }).click();
    });

    expect(screen.getByTestId("token").textContent).toBe("tok123");
    expect(screen.getByTestId("agencyId").textContent).toBe("ag-uuid");
    expect(screen.getByTestId("role").textContent).toBe("admin");
    expect(screen.getByTestId("email").textContent).toBe("a@b.com");

    expect(localStorage.getItem("auth_token")).toBe("tok123");
    expect(localStorage.getItem("auth_agency_id")).toBe("ag-uuid");
    expect(localStorage.getItem("auth_role")).toBe("admin");
    expect(localStorage.getItem("auth_email")).toBe("a@b.com");
  });

  it("logout clears state and localStorage", async () => {
    localStorage.setItem("auth_token", "tok");
    localStorage.setItem("auth_agency_id", "ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "a@b.com");

    render(<AuthProvider><TestConsumer /></AuthProvider>);

    await act(async () => {
      screen.getByRole("button", { name: "logout" }).click();
    });

    expect(screen.getByTestId("token").textContent).toBe("none");
    expect(localStorage.getItem("auth_token")).toBeNull();
    expect(localStorage.getItem("auth_agency_id")).toBeNull();
  });

  it("login stores email from the API response, not the caller argument", async () => {
    const body = {
      ...loginBody,
      user: { ...loginBody.user, email: "normalised@example.com" },
    };
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(200, body)));

    render(<AuthProvider><TestConsumer /></AuthProvider>);

    await act(async () => {
      screen.getByRole("button", { name: "login" }).click();
    });

    expect(screen.getByTestId("email").textContent).toBe("normalised@example.com");
    expect(localStorage.getItem("auth_email")).toBe("normalised@example.com");
  });

  it("auth:logout event triggers logout", async () => {
    localStorage.setItem("auth_token", "tok");
    localStorage.setItem("auth_agency_id", "ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "a@b.com");

    render(<AuthProvider><TestConsumer /></AuthProvider>);
    expect(screen.getByTestId("token").textContent).toBe("tok");

    await act(async () => {
      window.dispatchEvent(new Event("auth:logout"));
    });

    expect(screen.getByTestId("token").textContent).toBe("none");
  });
});
