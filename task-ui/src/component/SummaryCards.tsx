import { useMemo } from "react";
import type { Task } from "../types/api";

export type StatusFilter = "all" | Task["status"];

const CARDS: { label: string; value: StatusFilter }[] = [
  { label: "All",         value: "all" },
  { label: "Todo",        value: "todo" },
  { label: "In Progress", value: "in_progress" },
  { label: "Done",        value: "done" },
];

export function SummaryCards({ tasks, activeFilter, onFilterChange }: {
  tasks: Task[];
  activeFilter: StatusFilter;
  onFilterChange: (filter: StatusFilter) => void;
}) {
  const counts = useMemo(() => ({
    all:         tasks.length,
    todo:        tasks.filter((t) => t.status === "todo").length,
    in_progress: tasks.filter((t) => t.status === "in_progress").length,
    done:        tasks.filter((t) => t.status === "done").length,
  }), [tasks]);

  return (
    <div className="summary-cards">
      {CARDS.map(({ label, value }) => (
        <button
          key={value}
          className={`summary-card${activeFilter === value ? " summary-card--active" : ""}`}
          aria-pressed={activeFilter === value}
          onClick={() => onFilterChange(value)}
        >
          <span className="summary-card__count">{counts[value]}</span>
          <span className="summary-card__label">{label}</span>
        </button>
      ))}
    </div>
  );
}
