import type { Task } from "../types/api";

export type Priority = Task["priority"];

const priorityClass: Record<Priority, string> = {
  high:   "badge badge-red",
  medium: "badge badge-yellow",
  low:    "badge badge-gray",
};

const priorityLabel: Record<Priority, string> = {
  high:   "High",
  medium: "Medium",
  low:    "Low",
};

export function PriorityBadge({ priority }: { priority: Priority }) {
  return (
    <span className={priorityClass[priority] ?? "badge badge-gray"}>
      {priorityLabel[priority] ?? priority}
    </span>
  );
}
