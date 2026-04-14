import type { Task } from "../types/api";
import { StatusBadge } from "./StatusBadge";

export function TaskRow({ task }: { task: Task }) {
  return (
    <>
      <td>{task.title}</td>

      <td>
        <StatusBadge status={task.status as any} />
      </td>

      <td>{task.priority}</td>
    </>
  );
}