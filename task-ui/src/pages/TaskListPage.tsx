import { useState, useMemo } from "react";
import { Box, Button, CircularProgress, Typography } from "@mui/material";
import { useTasks } from "../hooks/useTasks";
import { TaskRow } from "../component/TaskRow";
import type { Task } from "../types/api";

type StatusFilter = "all" | Task["status"];

export function TaskListPage({ agencyId }: { agencyId: string }) {
  const { data: tasks, isLoading, isError, error, refetch } = useTasks(agencyId);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");

  const visible = useMemo(
    () => statusFilter === "all"
      ? (tasks ?? [])
      : (tasks ?? []).filter((t) => t.status === statusFilter),
    [tasks, statusFilter]
  );

  if (isLoading) {
    return (
      <Box role="status" sx={{ display: "flex", alignItems: "center", gap: 1, p: 4 }}>
        <CircularProgress size={20} aria-hidden="true" />
        <Typography>Loading tasks…</Typography>
      </Box>
    );
  }

  if (isError) {
    return (
      <Box role="alert" sx={{ display: "flex", flexDirection: "column", alignItems: "flex-start", gap: 1, p: 4 }}>
        <Typography sx={{ fontWeight: 600 }}>Failed to load tasks.</Typography>
        <Typography sx={{ color: "text.secondary" }}>
          {error instanceof Error ? error.message : "Something went wrong. Please try again."}
        </Typography>
        <Button variant="outlined" onClick={() => refetch()}>Retry</Button>
      </Box>
    );
  }

  return (
    <div>
      <div className="task-filters">
        <label htmlFor="status-filter" className="create-form__label">Status</label>
        <select
          id="status-filter"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
          className="create-form__select"
        >
          <option value="all">All</option>
          <option value="todo">Todo</option>
          <option value="in_progress">In Progress</option>
          <option value="done">Done</option>
        </select>
      </div>

      {visible.length === 0 ? (
        <p>No tasks found.</p>
      ) : (
        <table className="task-table">
          <tbody>
            {visible.map((task) => (
              <tr key={task.id}>
                <TaskRow task={task} />
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
