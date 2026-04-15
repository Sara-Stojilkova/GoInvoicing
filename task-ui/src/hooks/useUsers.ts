import { useQuery } from "@tanstack/react-query";
import { listUsers } from "../api/users";

export function useUsers(agencyId: string) {
  return useQuery({
    queryKey: ["users", agencyId],
    queryFn: () => listUsers(agencyId),
  });
}
