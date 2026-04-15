import { useState } from "react";
import { useCreateTask } from "../hooks/useTasks";

export function CreateTaskForm({ agencyId }: { agencyId: string }) {
  const { mutate, isPending, isError, error } = useCreateTask(agencyId);

  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState("medium");

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!title.trim()) return;

    mutate({
      title,
      priority,
      agency_id: agencyId,
    });

    setTitle("");
    setPriority("medium");
  };

  return (
    <form onSubmit={handleSubmit}>
      <label>
        Title
        <input name="title" value={title} onChange={(e) => setTitle(e.target.value)}/>
      </label>

      <label>
        Priority
        <select name="priority" value={priority} defaultValue="medium" onChange={(e) => setPriority(e.target.value)}>
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