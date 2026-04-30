import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { TagsInput } from "./TagsInput";

describe("TagsInput", () => {
  it("renders existing tags as chips", () => {
    render(<TagsInput value={["bug", "urgent"]} onChange={vi.fn()} />);
    expect(screen.getByText("bug")).toBeInTheDocument();
    expect(screen.getByText("urgent")).toBeInTheDocument();
  });

  it("renders an input for adding new tags", () => {
    render(<TagsInput value={[]} onChange={vi.fn()} />);
    expect(screen.getByRole("textbox")).toBeInTheDocument();
  });

  it("calls onChange with new tag when Enter is pressed", () => {
    const onChange = vi.fn();
    render(<TagsInput value={["bug"]} onChange={onChange} />);
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "urgent" } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith(["bug", "urgent"]);
  });

  it("trims whitespace from new tags", () => {
    const onChange = vi.fn();
    render(<TagsInput value={[]} onChange={onChange} />);
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "  bug  " } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith(["bug"]);
  });

  it("ignores empty input on Enter", () => {
    const onChange = vi.fn();
    render(<TagsInput value={[]} onChange={onChange} />);
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "   " } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).not.toHaveBeenCalled();
  });

  it("does not add duplicate tags", () => {
    const onChange = vi.fn();
    render(<TagsInput value={["bug"]} onChange={onChange} />);
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "bug" } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).not.toHaveBeenCalled();
  });

  it("removes a tag when its remove button is clicked", () => {
    const onChange = vi.fn();
    render(<TagsInput value={["bug", "urgent"]} onChange={onChange} />);
    fireEvent.click(screen.getByLabelText("Remove bug"));
    expect(onChange).toHaveBeenCalledWith(["urgent"]);
  });

  it("clears the input after adding a tag", () => {
    render(<TagsInput value={[]} onChange={vi.fn()} />);
    const input = screen.getByRole("textbox") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "bug" } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(input.value).toBe("");
  });

  it("is disabled when disabled prop is true", () => {
    render(<TagsInput value={["bug"]} onChange={vi.fn()} disabled />);
    expect(screen.getByRole("textbox")).toBeDisabled();
    expect(screen.queryByLabelText("Remove bug")).not.toBeInTheDocument();
  });
});
