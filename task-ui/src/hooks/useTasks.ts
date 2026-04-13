import { useQuery, useMutation } from "@tanstack/react-query";
import { listTasks } from "../api/tasks";

export function useCreateTask(_agencyId: string): ReturnType<typeof useMutation> {
  throw new Error("not implemented");
}

export function useTasks(agencyId: string) {
  return useQuery({
    queryKey: ["tasks", agencyId],
    queryFn: () => listTasks(agencyId),
  });
}