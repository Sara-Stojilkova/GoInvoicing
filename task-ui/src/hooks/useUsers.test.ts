// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useUsers } from "./useUsers";
import * as usersApi from "../api/users";
import type { User } from "../types/api";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";

const users: User[] = [
  {
    id: "b2c3d4e5-0000-0000-0000-000000000002",
    name: "Alice",
    email: "alice@acme.com",
    role: "admin",
    agency_id: agencyId,
    created_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "c3d4e5f6-0000-0000-0000-000000000003",
    name: "Bob",
    email: "bob@acme.com",
    role: "member",
    agency_id: agencyId,
    created_at: "2024-01-02T00:00:00Z",
  },
];

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("useUsers", () => {
  it("returns the list of users on success", async () => {
    vi.spyOn(usersApi, "listUsers").mockResolvedValue(users);

    const { result } = renderHook(() => useUsers(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(users);
  });

  it("calls listUsers with the given agencyId", async () => {
    const spy = vi.spyOn(usersApi, "listUsers").mockResolvedValue(users);

    const { result } = renderHook(() => useUsers(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(spy).toHaveBeenCalledWith(agencyId);
  });

  it("is in a loading state initially", () => {
    vi.spyOn(usersApi, "listUsers").mockReturnValue(new Promise(() => {}));

    const { result } = renderHook(() => useUsers(agencyId), { wrapper });

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
  });

  it("returns an empty array when the agency has no users", async () => {
    vi.spyOn(usersApi, "listUsers").mockResolvedValue([]);

    const { result } = renderHook(() => useUsers(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual([]);
  });

  it("sets isError when listUsers rejects", async () => {
    vi.spyOn(usersApi, "listUsers").mockRejectedValue(new Error("network error"));

    const { result } = renderHook(() => useUsers(agencyId), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
