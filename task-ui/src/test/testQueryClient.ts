import { QueryClient, type QueryObserverOptions } from "@tanstack/react-query";

export function createTestQueryClient(queryOverrides?: Partial<QueryObserverOptions>) {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
        gcTime: 0,
        refetchOnWindowFocus: false,
        ...queryOverrides,
      },
      mutations: {
        retry: false,
      },
    },
  });
}