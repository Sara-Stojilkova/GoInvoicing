import type { Task } from "../types/api";
import { StatusBadge } from "./StatusBadge";
import { useCompleteTask } from "../hooks/useTasks";

export function TaskRow({ task }: { task: Task }) {

  const { mutate, isPending } = useCompleteTask(task.agency_id);
  const isDone = task.status === "done";

  return (
    <>
      <td>{task.title}</td>
      <td><StatusBadge status={task.status} /></td>
      <td>{task.priority}</td>
      <td>{!isDone && (<button onClick={() => mutate(task.id)} disabled={isPending}>Complete</button>)}</td>
    </>
  );
}