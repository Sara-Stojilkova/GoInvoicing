import type { Task } from "../types/api";

export function useTasks(_agencyId: string): { data: Task[] | undefined; isLoading: boolean; isError: boolean } {
  throw new Error("not implemented");
}
