// @vitest-environment jsdom
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusBadge } from "./StatusBadge";

describe("StatusBadge", () => {
  it.each([
    ["todo", "Todo", "badge-gray"],
    ["in_progress", "In Progress", "badge-yellow"],
    ["done", "Done", "badge-green"],
  ])(
    "renders correct label and color for %s",
    (status, label, className) => {
      render(<StatusBadge status={status as any} />);

      const el = screen.getByText(label);
      expect(el).toBeInTheDocument();
      expect(el.closest("span")?.className).toContain(className);
    }
  );
});