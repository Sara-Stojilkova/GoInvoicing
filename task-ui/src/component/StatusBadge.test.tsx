// @vitest-environment jsdom
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusBadge } from "./StatusBadge";


describe("StatusBadge", () => {
    it("renders the correct status badge for todo", async () => {
      render(<StatusBadge status="todo" />);
      expect(screen.getByText("Todo")).toBeInTheDocument();
    });

    it("renders the correct color for todo", () => {
    render(<StatusBadge status="todo" />);
    expect(screen.getByText("Todo").closest("span")?.className).toContain("badge-gray");
    });

    it("renders the correct label for in_progress", () => {
    render(<StatusBadge status="in_progress" />);
    expect(screen.getByText("In Progress")).toBeInTheDocument();
    });

    it("renders the correct color for in_progress", () => {
    render(<StatusBadge status="in_progress" />);
    expect(screen.getByText("In Progress").closest("span")?.className).toContain("badge-yellow");
    });

    it("renders the correct label for done", () => {
    render(<StatusBadge status="done" />);
    expect(screen.getByText("Done")).toBeInTheDocument();
    });

    it("renders the correct color for done", () => {
    render(<StatusBadge status="done" />);
    expect(screen.getByText("Done").closest("span")?.className).toContain("badge-green");
    });
});