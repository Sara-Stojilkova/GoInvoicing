import { Box, Button, CircularProgress, Typography } from "@mui/material";
import { useTasks } from "../hooks/useTasks";
import { TaskRow } from "../component/TaskRow";
import { CreateTaskForm } from "../component/CreateTaskForm";

export function TaskListPage({ agencyId }: { agencyId: string }) {
  const { data: tasks, isLoading, isError, error, refetch } = useTasks(agencyId);

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

  if (!tasks || tasks.length === 0) {
    return (
      <>
        <p>No tasks found for this agency.</p>
        <CreateTaskForm agencyId={agencyId} />
      </>
    );
  }

  return (
    <>
      <table className="task-table">
        <tbody>
          {tasks.map((task) => (
            <tr key={task.id}>
              <TaskRow task={task} />
            </tr>
          ))}
        </tbody>
      </table>
      <CreateTaskForm agencyId={agencyId} />
    </>
  );
}
