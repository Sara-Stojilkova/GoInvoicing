import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { Header } from "./Header";
import { createAuthPageWrapper } from "../test/wrapper";

beforeEach(() => {
  localStorage.clear();
});

function renderHeader(initialEntries = ["/"]) {
  const Wrapper = createAuthPageWrapper(initialEntries);
  render(
    <Routes>
      <Route path="*" element={<Header />} />
      <Route path="/login" element={<div>login page</div>} />
    </Routes>,
    { wrapper: Wrapper }
  );
}

describe("Header", () => {
  it("shows Sign in button when not logged in", () => {
    renderHeader();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("shows the user email when logged in", () => {
    localStorage.setItem("auth_token", "tok");
    localStorage.setItem("auth_agency_id", "ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "jane@example.com");
    renderHeader();
    expect(screen.getByText("jane@example.com")).toBeInTheDocument();
  });

  it("shows Sign out button when logged in", () => {
    localStorage.setItem("auth_token", "tok");
    localStorage.setItem("auth_agency_id", "ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "jane@example.com");
    renderHeader();
    expect(screen.getByRole("button", { name: /sign out/i })).toBeInTheDocument();
  });

  it("navigates to /login when Sign out is clicked", async () => {
    localStorage.setItem("auth_token", "tok");
    localStorage.setItem("auth_agency_id", "ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "jane@example.com");

    const Wrapper = createAuthPageWrapper(["/"]);
    render(
      <Routes>
        <Route path="/" element={<Header />} />
        <Route path="/login" element={<div>login page</div>} />
      </Routes>,
      { wrapper: Wrapper }
    );

    await userEvent.click(screen.getByRole("button", { name: /sign out/i }));
    expect(screen.getByText("login page")).toBeInTheDocument();
  });
});
