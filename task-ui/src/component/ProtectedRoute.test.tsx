import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { ProtectedRoute } from "./ProtectedRoute";
import { createAuthPageWrapper } from "../test/wrapper";

beforeEach(() => {
  localStorage.clear();
});

describe("ProtectedRoute", () => {
  it("redirects to /login when no token", () => {
    const Wrapper = createAuthPageWrapper(["/"]);
    render(
      <Routes>
        <Route path="/login" element={<div>login page</div>} />
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<div>protected content</div>} />
        </Route>
      </Routes>,
      { wrapper: Wrapper }
    );
    expect(screen.getByText("login page")).toBeInTheDocument();
    expect(screen.queryByText("protected content")).not.toBeInTheDocument();
  });

  it("renders protected content when token is present", () => {
    localStorage.setItem("auth_token", "tok");
    localStorage.setItem("auth_agency_id", "ag");
    localStorage.setItem("auth_role", "admin");
    localStorage.setItem("auth_email", "a@b.com");

    const Wrapper = createAuthPageWrapper(["/"]);
    render(
      <Routes>
        <Route path="/login" element={<div>login page</div>} />
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<div>protected content</div>} />
        </Route>
      </Routes>,
      { wrapper: Wrapper }
    );
    expect(screen.getByText("protected content")).toBeInTheDocument();
    expect(screen.queryByText("login page")).not.toBeInTheDocument();
  });
});
