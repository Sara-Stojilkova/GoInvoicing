import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useAgency } from "./useAgency";
import * as agenciesApi from "../api/agencies";
import type { Agency } from "../types/api";
import { createWrapper } from "../test/wrapper";

const agency: Agency = {
  id: "a1b2c3d4-0000-0000-0000-000000000001",
  name: "Acme Corp",
  created_at: "2024-01-01T00:00:00Z",
};

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("useAgency", () => {
  it("returns the agency on success", async () => {
    vi.spyOn(agenciesApi, "getAgency").mockResolvedValue(agency);

    const { result } = renderHook(() => useAgency(agency.id), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(agency);
  });

  it("calls getAgency with the given agencyId", async () => {
    const spy = vi.spyOn(agenciesApi, "getAgency").mockResolvedValue(agency);

    const { result } = renderHook(() => useAgency(agency.id), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(spy).toHaveBeenCalledWith(agency.id);
  });

  it("is in a loading state initially", () => {
    vi.spyOn(agenciesApi, "getAgency").mockReturnValue(new Promise(() => {}));

    const { result } = renderHook(() => useAgency(agency.id), { wrapper: createWrapper() });

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
  });

  it("sets isError when getAgency rejects", async () => {
    vi.spyOn(agenciesApi, "getAgency").mockRejectedValue(new Error("network error"));

    const { result } = renderHook(() => useAgency(agency.id), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
