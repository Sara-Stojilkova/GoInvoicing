import { useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Box, CircularProgress, Typography } from "@mui/material";
import { useTask, useCompleteTask, useAssignTask, useSetInProgress, useUpdateDueDate, useUpdateDescription } from "../hooks/useTasks";
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
  const { mutate: complete } = useCompleteTask(agencyId);
  const { mutate: setInProgress } = useSetInProgress(agencyId);
  const { mutate: assign } = useAssignTask(agencyId);
  const { mutate: updateDueDate } = useUpdateDueDate(agencyId);
  const { mutate: updateDescription } = useUpdateDescription(agencyId);
  const dueDateRef = useRef<HTMLInputElement>(null);
  const [editingDescription, setEditingDescription] = useState(false);

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
          <div className="detail-date-field">
            {task.due_date
              ? <span>{formatDate(task.due_date)}</span>
              : <span className="detail-field__empty">No due date</span>}
            <button
              type="button"
              className="detail-date-btn"
              aria-label="Edit due date"
              onClick={() => dueDateRef.current?.showPicker?.()}
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                <line x1="16" y1="2" x2="16" y2="6"/>
                <line x1="8" y1="2" x2="8" y2="6"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
              </svg>
            </button>
            <input
              key={task.due_date}
              ref={dueDateRef}
              type="date"
              aria-label="due-date-input"
              className="detail-date-input--hidden"
              defaultValue={task.due_date ? task.due_date.split("T")[0] : ""}
              onBlur={(e) => {
                const value = e.target.value || null;
                updateDueDate({ taskId: task.id, dueDate: value });
              }}
            />
          </div>
        </Field>
        <Field label="Completed">
          {task.completed_at
            ? formatDate(task.completed_at)
            : <span className="detail-field__empty">Not completed</span>}
        </Field>
        <Field label="Description">
          <button
            type="button"
            className="detail-date-btn"
            aria-label="Edit description"
            onClick={() => setEditingDescription(true)}
          >
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
          </button>
          {editingDescription ? (
            <textarea
              key={task.description}
              aria-label="Description"
              className="detail-description-input"
              defaultValue={task.description ?? ""}
              autoFocus
              onBlur={(e) => {
                const value = e.target.value || null;
                updateDescription({ taskId: task.id, description: value });
                setEditingDescription(false);
              }}
            />
          ) : (
            task.description
              ? task.description
              : <span className="detail-field__empty">No description</span>
          )}
        </Field>
      </dl>
      <div className="detail-actions">
        <label htmlFor="status-action" className="detail-field__label">Change status</label>
        <select
          id="status-action"
          aria-label="Change status"
          className="create-form__select detail-actions__select"
          value=""
          onChange={(e) => {
            if (e.target.value === "complete") complete(task.id);
            if (e.target.value === "in_progress") setInProgress(task.id);
          }}
        >
          <option value="" disabled hidden>Change status…</option>
          <option value="complete" disabled={task.status === "done"}>Complete</option>
          <option value="in_progress" disabled={task.status === "in_progress"}>Set In Progress</option>
        </select>
      </div>
    </div>
  );
}
