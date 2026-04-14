import { useRef } from "react";
import { useCreateTask } from "../hooks/useTasks";

export function CreateTaskForm({ agencyId }: { agencyId: string }) {
  const { mutate, isPending, isError, error } = useCreateTask(agencyId);

  const titleRef = useRef<HTMLInputElement>(null);
  const priorityRef = useRef<HTMLSelectElement>(null);

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const title = titleRef.current?.value || "";
    const priority = priorityRef.current?.value || "";

    if (!title.trim()) return;

    mutate({
      title,
      priority,
      agency_id: agencyId, // ✅ REQUIRED here
    });

    if (titleRef.current) titleRef.current.value = "";
    if (priorityRef.current) priorityRef.current.value = "medium";
  };

  return (
    <form onSubmit={handleSubmit}>
      <label>
        Title
        <input ref={titleRef} name="title" />
      </label>

      <label>
        Priority
        <select ref={priorityRef} name="priority" defaultValue="medium">
          <option value="low">low</option>
          <option value="medium">medium</option>
          <option value="high">high</option>
        </select>
      </label>

      {isError && <p>{error?.message}</p>}

      <button type="submit" disabled={isPending}>
        Create
      </button>
    </form>
  );
}