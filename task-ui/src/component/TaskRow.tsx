import { Link } from "react-router-dom";
import type { Task } from "../types/api";
import { StatusBadge } from "./StatusBadge";
import { PriorityBadge } from "./PriorityBadge";
import { useCompleteTask } from "../hooks/useTasks";

export function TaskRow({ task }: { task: Task }) {

  const { mutate, isPending } = useCompleteTask(task.agency_id);
  const isDone = task.status === "done";

  return (
    <>
      <td><Link to={`/tasks/${task.id}`}>{task.title}</Link></td>
      <td><StatusBadge status={task.status} /></td>
      <td><PriorityBadge priority={task.priority} /></td>
      <td>
        <button className="btn-complete" onClick={() => mutate(task.id)} disabled={isPending || isDone}>
          {isPending ? (
            <span className="spinner" aria-label="loading" />
          ) : (
            "Complete"
          )}
        </button>
      </td>
    </>
  );
}