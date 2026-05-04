import { useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Box, CircularProgress, Typography } from "@mui/material";
import { useTask, useCompleteTask, useAssignTask, useSetInProgress, useUpdateDueDate, useUpdateDescription, useUpdateTags } from "../hooks/useTasks";
import { useUsers } from "../hooks/useUsers";
import { useAgency } from "../hooks/useAgency";
import { StatusBadge } from "../component/StatusBadge";
import { PriorityBadge } from "../component/PriorityBadge";
import { TagsInput } from "../component/TagsInput";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric", month: "short", day: "numeric",
  });
}

// Sidebar property row
function Prop({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="task-prop">
      <dt className="task-prop__label">{label}</dt>
      <dd className="task-prop__value">{children}</dd>
    </div>
  );
}

const ChevronLeftIcon = () => (
  <svg className="back-link__chevron" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <polyline points="15 18 9 12 15 6" />
  </svg>
);

// Icons
const CalendarIcon = () => (
  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
    <line x1="16" y1="2" x2="16" y2="6"/>
    <line x1="8" y1="2" x2="8" y2="6"/>
    <line x1="3" y1="10" x2="21" y2="10"/>
  </svg>
);

const PencilIcon = () => (
  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
  </svg>
);

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
  const { mutate: updateTags } = useUpdateTags(agencyId);
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
        <Link to="/" className="back-link"><ChevronLeftIcon />Tasks</Link>
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
  const assignedUser = users?.find(u => u.id === task.assigned_to);

  return (
    <div className="task-detail">
      <header className="task-detail__header">
        <div className="task-detail__header-top">
          <Link to="/" className="back-link"><ChevronLeftIcon />Tasks</Link>
          <h1>{task.title}</h1>
        </div>
        <div className="task-detail__header-badges">
          <StatusBadge status={task.status} />
          <PriorityBadge priority={task.priority} />
        </div>
      </header>

      <div className="task-detail__body">

        {/* ── Left: description ── */}
        <section className="task-detail__main">
          <p className="task-section-label">Description</p>
          <div className="task-detail__desc-area">
            <button
              type="button"
              className="detail-date-btn"
              aria-label="Edit description"
              onClick={() => setEditingDescription(true)}
            >
              <PencilIcon />
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
                ? <p className="task-detail__desc-text">{task.description}</p>
                : <span className="detail-field__empty">No description — click the pencil to add one.</span>
            )}
          </div>
        </section>

        {/* ── Right: properties sidebar ── */}
        <aside className="task-detail__sidebar">
          <dl className="task-props">
            <Prop label="Assignee">
              <select
                id="assignee"
                aria-label="Assignee"
                className="task-prop__select"
                value={task.assigned_to ?? ""}
                onChange={(e) => {
                  assign({ taskId: task.id, assigneeId: e.target.value || null, assigneeAgencyId: agencyId });
                }}
              >
                <option value="">Unassigned</option>
                {users?.map((u) => (
                  <option key={u.id} value={u.id}>{u.full_name}</option>
                ))}
              </select>
            </Prop>

            <Prop label="Due date">
              <div className="task-prop__date">
                {task.due_date
                  ? <span>{formatDate(task.due_date)}</span>
                  : <span className="detail-field__empty">No due date</span>}
                <button
                  type="button"
                  className="task-prop__icon-btn"
                  aria-label="Edit due date"
                  onClick={() => dueDateRef.current?.showPicker?.()}
                >
                  <CalendarIcon />
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
            </Prop>

            <Prop label="Tags">
              <TagsInput
                value={task.tags ?? []}
                onChange={(tags) => updateTags({ taskId: task.id, tags })}
              />
            </Prop>

            <Prop label="Created">{formatDate(task.created_at)}</Prop>

            <Prop label="Completed">
              {task.completed_at
                ? formatDate(task.completed_at)
                : <span className="detail-field__empty">Not completed</span>}
            </Prop>

            <Prop label="Agency">{agencyName}</Prop>

            <Prop label="ID">
              <span className="task-prop__id">{task.id}</span>
            </Prop>
          </dl>

          {/* Status action */}
          <div className="task-detail__actions">
            <p className="task-section-label">Change status</p>
            <select
              id="status-action"
              aria-label="Change status"
              className="task-prop__select task-prop__select--full"
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

          {/* Assigned member card */}
          {assignedUser && (
            <div className="task-detail__assignee-card">
              <div className="task-detail__avatar">
                {(assignedUser.full_name || assignedUser.email || "?").charAt(0).toUpperCase()}
              </div>
              <p className="task-detail__assignee-email">{assignedUser.email}</p>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}
