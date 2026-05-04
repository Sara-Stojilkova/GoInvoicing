import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { supabase } from "../lib/supabase";

/**
 * Subscribes to Supabase realtime changes on the tasks table scoped to the
 * current agency. Any INSERT, UPDATE, or DELETE invalidates the React Query
 * cache so the list refreshes automatically.
 *
 * Prerequisite: the tasks table must be added to the supabase_realtime
 * publication. Run once in Supabase SQL editor:
 *   ALTER PUBLICATION supabase_realtime ADD TABLE tasks;
 */
export function useTasksRealtime(agencyId: string) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!agencyId) return;

    const channel = supabase
      .channel(`tasks:agency:${agencyId}`)
      .on(
        "postgres_changes",
        {
          event: "*",
          schema: "public",
          table: "tasks",
          filter: `agency_id=eq.${agencyId}`,
        },
        (payload: any) => {
          console.log(
          "REALTIME EVENT:",
          payload.eventType,
          payload.new?.agency_id
        );

        queryClient.invalidateQueries({ queryKey: ["tasks", agencyId] });
        }
      )
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  }, [agencyId, queryClient]);
}
