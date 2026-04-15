import { Link, useParams } from "react-router-dom";
import { Box, CircularProgress, Typography } from "@mui/material";
import { useTask } from "../hooks/useTasks";
import { StatusBadge } from "../component/StatusBadge";

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
        <Link to="/" className="back-link">← Back to list</Link>
        <Typography sx={{ fontWeight: 600, mt: 2 }}>
          {(error as { status?: number })?.status === 404
            ? "Task not found."
            : "Failed to load task."}
        </Typography>
      </Box>
    );
  }

  if (!task) return null;

  return (
    <div className="page">
      <Link to="/" className="back-link">← Back to list</Link>
      <h1>{task.title}</h1>
      <dl className="detail-grid">
        <Field label="Status"><StatusBadge status={task.status} /></Field>
        <Field label="Priority">{task.priority}</Field>
        <Field label="ID">{task.id}</Field>
        <Field label="Agency">{task.agency_id}</Field>
        <Field label="Created">{formatDate(task.created_at)}</Field>
        <Field label="Assignee">
          {task.assignee_id ?? <span className="detail-field__empty">Unassigned</span>}
        </Field>
        <Field label="Due date">
          {task.due_date
            ? formatDate(task.due_date)
            : <span className="detail-field__empty">No due date</span>}
        </Field>
        {task.completed_at && (
          <Field label="Completed">{formatDate(task.completed_at)}</Field>
        )}
        {task.description && (
          <Field label="Description">{task.description}</Field>
        )}
      </dl>
    </div>
  );
}
