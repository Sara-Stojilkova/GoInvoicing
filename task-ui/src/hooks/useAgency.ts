import { useQuery } from "@tanstack/react-query";
import { getAgency } from "../api/agencies";

export function useAgency(agencyId: string) {
  return useQuery({
    queryKey: ["agencies", agencyId],
    queryFn: () => getAgency(agencyId),
  });
}
