import { useQuery } from "@tanstack/react-query";
import { listTasks } from "../api/tasks";

export function useTasks(agencyId: string) {
  return useQuery({
    queryKey: ["tasks", agencyId],
    queryFn: () => listTasks(agencyId),
  });
}