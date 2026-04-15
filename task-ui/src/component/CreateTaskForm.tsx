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

    mutate(
      {
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
    <form onSubmit={handleSubmit} className="create-form">
      <div className="create-form__field">
        <label className="create-form__label" htmlFor="cf-title">Title</label>
        <input
          id="cf-title"
          name="title"
          value={title}
          onChange={(e) => { setTitle(e.target.value); if (errors.title) setErrors({}); }}
          className="create-form__input"
          placeholder="Task title"
        />
        {errors.title && <p className="create-form__error">{errors.title}</p>}
      </div>

      <div className="create-form__field">
        <label className="create-form__label" htmlFor="cf-priority">Priority</label>
        <select
          id="cf-priority"
          name="priority"
          value={priority}
          onChange={(e) => setPriority(e.target.value)}
          className="create-form__select"
        >
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
      </div>

      <div className="create-form__field">
        <label className="create-form__label" htmlFor="cf-description">Description</label>
        <input
          id="cf-description"
          name="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="create-form__input"
          placeholder="Optional"
        />
      </div>

      <div className="create-form__field">
        <label className="create-form__label" htmlFor="cf-assignee">Assignee</label>
        <select
          id="cf-assignee"
          name="assignee_id"
          value={assigneeId}
          disabled={usersLoading}
          onChange={(e) => setAssigneeId(e.target.value)}
          className="create-form__select"
        >
          <option value="">Unassigned</option>
          {(users ?? []).map((user) => (
            <option key={user.id} value={user.id}>{user.name}</option>
          ))}
        </select>
      </div>

      <div className="create-form__field">
        <label className="create-form__label" htmlFor="cf-due-date">Due Date</label>
        <input
          id="cf-due-date"
          type="date"
          name="due_date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          className="create-form__input"
        />
      </div>

      {isError && <p className="create-form__api-error">{error?.message}</p>}

      <button type="submit" disabled={isPending} className="create-form__submit">
        Create
      </button>
    </form>
  );
}
