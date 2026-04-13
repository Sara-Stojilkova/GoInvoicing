import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listTasks, createTask, completeTask } from "../api/tasks";

export function useCompleteTask(_agencyId: string): ReturnType<typeof useMutation> {
  throw new Error("not implemented");
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