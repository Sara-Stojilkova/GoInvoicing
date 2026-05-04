import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { LoginPage } from "./LoginPage";
import { createAuthPageWrapper } from "../test/wrapper";

function makeFetchResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

const loginBody = {
  access_token: "tok",
  refresh_token: "ref",
  expires_in: 3600,
  token_type: "bearer",
  user: {
    email: "a@b.com",
    app_metadata: { agency_id: "ag-1", role: "admin" },
  },
};

beforeEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

function renderLogin(initialEntries = ["/login"]) {
  const Wrapper = createAuthPageWrapper(initialEntries);
  render(
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>home page</div>} />
    </Routes>,
    { wrapper: Wrapper }
  );
}

describe("LoginPage", () => {
  it("renders email and password fields and a submit button", () => {
    renderLogin();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("shows a link to the register page", () => {
    renderLogin();
    expect(screen.getByRole("link", { name: /register/i })).toBeInTheDocument();
  });

  it("shows success banner when ?registered=true is in the URL", () => {
    renderLogin(["/login?registered=true"]);
    expect(screen.getByText(/account created/i)).toBeInTheDocument();
  });

  it("redirects to / on successful login", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(200, loginBody)));
    renderLogin();

    await userEvent.type(screen.getByLabelText(/email/i), "a@b.com");
    await userEvent.type(screen.getByLabelText(/password/i), "pass");
    await userEvent.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => expect(screen.getByText("home page")).toBeInTheDocument());
  });

  it("shows an error message on invalid credentials (401)", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(401, { error: "invalid credentials" })));
    renderLogin();

    await userEvent.type(screen.getByLabelText(/email/i), "a@b.com");
    await userEvent.type(screen.getByLabelText(/password/i), "wrong");
    await userEvent.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => expect(screen.getByText(/invalid email or password/i)).toBeInTheDocument());
  });

  it("shows a generic error on server error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(makeFetchResponse(500, { error: "internal error" })));
    renderLogin();

    await userEvent.type(screen.getByLabelText(/email/i), "a@b.com");
    await userEvent.type(screen.getByLabelText(/password/i), "pass");
    await userEvent.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => expect(screen.getByText(/something went wrong/i)).toBeInTheDocument());
  });
});
