import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { PriorityBadge } from "./PriorityBadge";

describe("PriorityBadge", () => {
  it.each([
    ["high",   "High",   "badge-red"],
    ["medium", "Medium", "badge-yellow"],
    ["low",    "Low",    "badge-gray"],
  ])(
    "renders correct label and color for %s",
    (priority, label, className) => {
      render(<PriorityBadge priority={priority as any} />);

      const el = screen.getByText(label);
      expect(el).toBeInTheDocument();
      expect(el.closest("span")?.className).toContain(className);
    }
  );
});
