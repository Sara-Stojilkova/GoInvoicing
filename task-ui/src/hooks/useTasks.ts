import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listTasks, getTask, createTask, completeTask } from "../api/tasks";

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

export function useTasks(agencyId: string) {
  return useQuery({
    queryKey: ["tasks", agencyId],
    queryFn: () => listTasks(agencyId),
  });
}