import React from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { createTestQueryClient } from "./testQueryClient";

export function createWrapper() {
  const queryClient = createTestQueryClient();

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };
}