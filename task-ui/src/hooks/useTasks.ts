import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listTasks, getTask, createTask, completeTask, assignTask, unassignTask, setTaskInProgress } from "../api/tasks";

export function useTask(taskId: string | null, agencyId: string) {
  return useQuery({
    queryKey: ["tasks", agencyId, taskId],
    queryFn: () => getTask(taskId!, agencyId),
    enabled: taskId !== null,
  });
}

export function useCompleteTask(agencyId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: completeTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks", agencyId] });
    },
  });
}

export function useCreateTask(agencyId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks", agencyId] });
    },
  });
}

export function useSetInProgress(agencyId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: setTaskInProgress,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks", agencyId] });
    },
  });
}

export function useAssignTask(agencyId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, assigneeId, assigneeAgencyId }: { taskId: string; assigneeId: string | null; assigneeAgencyId: string }) =>
      assigneeId
        ? assignTask(taskId, { assignee_id: assigneeId, assignee_agency_id: assigneeAgencyId })
        : unassignTask(taskId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks", agencyId] });
    },
  });
}

export function useTasks(agencyId: string) {
  return useQuery({
    queryKey: ["tasks", agencyId],
    queryFn: () => listTasks(agencyId),
  });
}