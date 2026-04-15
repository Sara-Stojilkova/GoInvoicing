import { useState } from "react";
import { useCreateTask } from "../hooks/useTasks";
import { useUsers } from "../hooks/useUsers";

export function CreateTaskForm({ agencyId }: { agencyId: string }) {
  const { mutate, isPending, isError, error } = useCreateTask(agencyId);
  const { data: users, isLoading: usersLoading } = useUsers(agencyId);

  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState("medium");
  const [description, setDescription] = useState("");
  const [assigneeId, setAssigneeId] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [errors, setErrors] = useState<{ title?: string }>({});

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!title.trim()) {
      setErrors({ title: "Title is required" });
      return;
    }

    setErrors({});

    mutate({
      title,
      priority,
      agency_id: agencyId,
      description: description || undefined,
      assignee_id: assigneeId || undefined,
      due_date: dueDate || undefined,
    },
    {
      onSuccess: () => {
        setTitle("");
        setPriority("medium");
        setDescription("");
        setAssigneeId("");
        setDueDate("");
      },
    }
    );
  };

  return (
    <form onSubmit={handleSubmit}>
      <label>
        Title
        <input name="title" value={title} onChange={(e) => { setTitle(e.target.value); if (errors.title) setErrors({}); }}/>
      </label>
      {errors.title && <p>{errors.title}</p>}

      <label>
        Priority
        <select name="priority" value={priority} onChange={(e) => setPriority(e.target.value)}>
          <option value="low">low</option>
          <option value="medium">medium</option>
          <option value="high">high</option>
        </select>
      </label>

      <label>
        Description
        <input name="description" value={description} onChange={(e) => setDescription(e.target.value)} />
      </label>

      <label>
        Assignee
        <select name="assignee_id" value={assigneeId} disabled={usersLoading} onChange={(e) => setAssigneeId(e.target.value)} >
          <option value="">Unassigned</option>
          {(users ?? []).map((user) => (
            <option key={user.id} value={user.id}>{user.name}</option>
          ))}
        </select>
      </label>

      <label>
        Due Date
        <input type="date" name="due_date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
      </label>

      {isError && <p>{error?.message}</p>}

      <button type="submit" disabled={isPending}>
        Create
      </button>
    </form>
  );
}