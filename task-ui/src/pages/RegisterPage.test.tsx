import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { RegisterPage } from "./RegisterPage";
import { createAuthPageWrapper } from "../test/wrapper";

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

function renderRegister() {
  const Wrapper = createAuthPageWrapper(["/register"]);
  render(
    <Routes>
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/login" element={<div>login page</div>} />
    </Routes>,
    { wrapper: Wrapper }
  );
}

describe("RegisterPage", () => {
  it("renders full name, email, and password fields", () => {
    renderRegister();
    expect(screen.getByLabelText(/full name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it("shows agency name field when 'Create new' is selected (default)", () => {
    renderRegister();
    expect(screen.getByLabelText(/agency name/i)).toBeInTheDocument();
    expect(screen.queryByLabelText(/agency id/i)).not.toBeInTheDocument();
  });

  it("shows agency ID field when 'Join existing' is selected", async () => {
    renderRegister();
    await userEvent.click(screen.getByLabelText(/join existing/i));
    expect(screen.getByLabelText(/agency id/i)).toBeInTheDocument();
    expect(screen.queryByLabelText(/agency name/i)).not.toBeInTheDocument();
  });

  it("redirects to /login?registered=true on successful registration", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        makeFetchResponse(201, { user_id: "u1", agency_id: "ag-1", role: "admin", activated: true })
      )
    );
    renderRegister();

    await userEvent.type(screen.getByLabelText(/full name/i), "Jane Doe");
    await userEvent.type(screen.getByLabelText(/email/i), "jane@example.com");
    await userEvent.type(screen.getByLabelText(/password/i), "pass123");
    await userEvent.type(screen.getByLabelText(/agency name/i), "Acme Co");
    await userEvent.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(screen.getByText("login page")).toBeInTheDocument());
  });

  it("shows 'Email already registered' error on 409", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(409, { error: "email already registered" })));
    renderRegister();

    await userEvent.type(screen.getByLabelText(/full name/i), "Jane");
    await userEvent.type(screen.getByLabelText(/email/i), "jane@example.com");
    await userEvent.type(screen.getByLabelText(/password/i), "pass");
    await userEvent.type(screen.getByLabelText(/agency name/i), "Acme");
    await userEvent.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(screen.getByText(/email already registered/i)).toBeInTheDocument());
  });

  it("shows 'Agency not found' error on 404", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(404, { error: "agency not found" })));
    renderRegister();

    await userEvent.click(screen.getByLabelText(/join existing/i));
    await userEvent.type(screen.getByLabelText(/full name/i), "Bob");
    await userEvent.type(screen.getByLabelText(/email/i), "bob@example.com");
    await userEvent.type(screen.getByLabelText(/password/i), "pass");
    await userEvent.type(screen.getByLabelText(/agency id/i), "00000000-0000-0000-0000-000000000000");
    await userEvent.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(screen.getByText(/agency not found/i)).toBeInTheDocument());
  });

  it("shows a link to the login page", () => {
    renderRegister();
    expect(screen.getByRole("link", { name: /sign in/i })).toBeInTheDocument();
  });
});
