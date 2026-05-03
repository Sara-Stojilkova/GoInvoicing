import { Link } from "react-router-dom";
import type { Task, User } from "../types/api";
import { StatusBadge } from "./StatusBadge";
import { PriorityBadge } from "./PriorityBadge";
import { useCompleteTask } from "../hooks/useTasks";

const CheckIcon = () => (
  <svg width="11" height="11" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <polyline points="2 6 5 9 10 3" />
  </svg>
);

export function TaskRow({ task, users = [] }: { task: Task; users?: User[] }) {
  const { mutate, isPending } = useCompleteTask(task.agency_id);
  const isDone = task.status === "done";
  const assignee = users.find(u => u.id === task.assigned_to);

  return (
    <>
      <td>
        <div className="task-title-cell">
          <button
            className={`task-check-btn${isDone ? " task-check-btn--done" : ""}`}
            aria-label="Complete task"
            onClick={() => mutate(task.id)}
            disabled={isPending || isDone}
          >
            {isPending ? (
              <span className="spinner" aria-label="loading" />
            ) : isDone ? (
              <CheckIcon />
            ) : null}
          </button>
          <Link to={`/tasks/${task.id}`} className={isDone ? "task-title--done" : undefined}>{task.title}</Link>
          <StatusBadge status={task.status} />
          <PriorityBadge priority={task.priority} />
        </div>
      </td>
      <td className="task-table__assignee-col">
        {assignee && (
          <div className="task-row-avatar" data-tooltip={assignee.full_name}>
            {assignee.full_name ?? "".charAt(0).toUpperCase()}
          </div>
        )}
      </td>
    </>
  );
}