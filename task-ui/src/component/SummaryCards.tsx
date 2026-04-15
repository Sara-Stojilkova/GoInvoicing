import type { Task } from "../types/api";

export type StatusFilter = "all" | Task["status"];

export function SummaryCards(_props: {
  tasks: Task[];
  activeFilter: StatusFilter;
  onFilterChange: (filter: StatusFilter) => void;
}): never {
  throw new Error("not implemented");
}
