import { Link, useParams } from "react-router-dom";
import { Box, CircularProgress, Typography } from "@mui/material";
import { useTask, useCompleteTask, useAssignTask } from "../hooks/useTasks";
import { useUsers } from "../hooks/useUsers";
import { useAgency } from "../hooks/useAgency";
import { StatusBadge } from "../component/StatusBadge";
import { PriorityBadge } from "../component/PriorityBadge";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric", month: "short", day: "numeric",
  });
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="detail-field">
      <dt className="detail-field__label">{label}</dt>
      <dd className="detail-field__value">{children}</dd>
    </div>
  );
}

export function TaskDetailPage({ agencyId }: { agencyId: string }) {
  const { taskId = null } = useParams<{ taskId: string }>();
  const { data: task, isLoading, isError, error } = useTask(taskId, agencyId);
  const { data: users } = useUsers(agencyId);
  const { data: agency } = useAgency(agencyId);
  const { mutate: complete, isPending: isCompleting } = useCompleteTask(agencyId);
  const { mutate: assign } = useAssignTask(agencyId);

  if (isLoading) {
    return (
      <Box role="status" sx={{ display: "flex", alignItems: "center", gap: 1, p: 4 }}>
        <CircularProgress size={20} aria-hidden="true" />
        <Typography>Loading task…</Typography>
      </Box>
    );
  }

  if (isError) {
    return (
      <Box role="alert" sx={{ p: 4 }}>
        <Link to="/" className="back-link">Back to list</Link>
        <Typography sx={{ fontWeight: 600, mt: 2 }}>
          {(error as { status?: number })?.status === 404
            ? "Task not found."
            : "Failed to load task."}
        </Typography>
      </Box>
    );
  }

  if (!task) return null;

  const agencyName = agency?.name ?? task.agency_id;

  return (
    <div className="page">
      <Link to="/" className="back-link">Back to list</Link>
      <h1>{task.title}</h1>
      <dl className="detail-grid">
        <Field label="Status"><StatusBadge status={task.status} /></Field>
        <Field label="Priority"><PriorityBadge priority={task.priority} /></Field>
        <Field label="ID">{task.id}</Field>
        <Field label="Agency">{agencyName}</Field>
        <Field label="Created">{formatDate(task.created_at)}</Field>
        <Field label="Assignee">
          <select
            id="assignee"
            aria-label="Assignee"
            className="create-form__select"
            value={task.assignee_id ?? ""}
            onChange={(e) => {
              assign({ taskId: task.id, assigneeId: e.target.value || null, assigneeAgencyId: agencyId });
            }}
          >
            <option value="">Unassigned</option>
            {users?.map((u) => (
              <option key={u.id} value={u.id}>{u.name}</option>
            ))}
          </select>
        </Field>
        <Field label="Due date">
          {task.due_date
            ? formatDate(task.due_date)
            : <span className="detail-field__empty">No due date</span>}
        </Field>
        <Field label="Completed">
          {task.completed_at
            ? formatDate(task.completed_at)
            : <span className="detail-field__empty">Not completed</span>}
        </Field>
        <Field label="Description">
          {task.description
            ? task.description
            : <span className="detail-field__empty">No description</span>}
        </Field>
      </dl>
      {task.status !== "done" && (
        <div className="detail-actions">
          <button
            className="btn-complete"
            onClick={() => complete(task.id)}
            disabled={isCompleting}
          >
            {isCompleting ? <span className="spinner" aria-label="loading" /> : "Complete"}
          </button>
        </div>
      )}
    </div>
  );
}
