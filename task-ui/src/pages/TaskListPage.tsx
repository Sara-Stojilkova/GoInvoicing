import { useState, useMemo } from "react";
import { Box, Button, CircularProgress, Typography } from "@mui/material";
import { useTasks } from "../hooks/useTasks";
import { useUsers } from "../hooks/useUsers";
import { TaskRow } from "../component/TaskRow";
import { SummaryCards } from "../component/SummaryCards";
import type { StatusFilter } from "../component/SummaryCards";
import { CreateTaskForm } from "../component/CreateTaskForm";

export function TaskListPage({ agencyId }: { agencyId: string }) {
  const { data: tasks, isLoading, isError, error, refetch } = useTasks(agencyId);
  const { data: users = [] } = useUsers(agencyId);
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
    <div className="task-list-page">
      <div className="task-list-body">
        <div className="task-list-left">
          <section className="page-section">
            <SummaryCards
              tasks={tasks ?? []}
              activeFilter={statusFilter}
              onFilterChange={setStatusFilter}
            />
          </section>

          <section className="task-list-section">
            <h2 className="section-title">Task list</h2>
            {visible.length === 0 ? (
              <p className="empty-state">No tasks found.</p>
            ) : (
              <table className="task-table">
                <tbody>
                  {visible.map((task) => (
                    <tr key={task.id}>
                      <TaskRow task={task} users={users} />
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </section>
        </div>

        <aside className="task-form-aside">
          <h2 className="section-title">New task</h2>
          <CreateTaskForm agencyId={agencyId} />
        </aside>
      </div>
    </div>
  );
}

