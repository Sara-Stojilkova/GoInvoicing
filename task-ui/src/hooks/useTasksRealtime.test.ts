import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { createElement } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { useTasksRealtime } from "./useTasksRealtime";
import { createTestQueryClient } from "../test/testQueryClient";
import * as supabaseLib from "../lib/supabase";

const agencyId = "a1b2c3d4-0000-0000-0000-000000000001";
const otherAgencyId = "bbbbbbbb-0000-0000-0000-000000000002";

function makeChannel() {
  const channel = {
    on: vi.fn(),
    subscribe: vi.fn(),
  };
  channel.on.mockReturnValue(channel);
  channel.subscribe.mockReturnValue(channel);
  return channel;
}

function makeWrapper() {
  const queryClient = createTestQueryClient({ gcTime: Infinity });
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
  return { queryClient, Wrapper };
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("useTasksRealtime", () => {
  it("creates a channel scoped to the agency", () => {
    const channel = makeChannel();
    const channelSpy = vi.spyOn(supabaseLib.supabase, "channel").mockReturnValue(channel as any);
    vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { Wrapper } = makeWrapper();

    renderHook(() => useTasksRealtime(agencyId), { wrapper: Wrapper });

    expect(channelSpy).toHaveBeenCalledWith(`tasks:agency:${agencyId}`);
  });

  it("subscribes to postgres_changes on tasks filtered by agency_id", () => {
    const channel = makeChannel();
    vi.spyOn(supabaseLib.supabase, "channel").mockReturnValue(channel as any);
    vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { Wrapper } = makeWrapper();

    renderHook(() => useTasksRealtime(agencyId), { wrapper: Wrapper });

    expect(channel.on).toHaveBeenCalledWith(
      "postgres_changes",
      {
        event: "*",
        schema: "public",
        table: "tasks",
        filter: `agency_id=eq.${agencyId}`,
      },
      expect.any(Function)
    );
    expect(channel.subscribe).toHaveBeenCalled();
  });

  it("invalidates the tasks cache for the agency when a realtime event fires", () => {
    const channel = makeChannel();
    let realtimeCallback: (() => void) | null = null;
    channel.on.mockImplementation((_event: string, _filter: unknown, cb: () => void) => {
      realtimeCallback = cb;
      return channel;
    });
    vi.spyOn(supabaseLib.supabase, "channel").mockReturnValue(channel as any);
    vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { queryClient, Wrapper } = makeWrapper();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    renderHook(() => useTasksRealtime(agencyId), { wrapper: Wrapper });

    realtimeCallback!();

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["tasks", agencyId] });
  });

  it("does not create a channel when agencyId is empty", () => {
    const channelSpy = vi.spyOn(supabaseLib.supabase, "channel");
    vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { Wrapper } = makeWrapper();

    renderHook(() => useTasksRealtime(""), { wrapper: Wrapper });

    expect(channelSpy).not.toHaveBeenCalled();
  });

  it("removes the channel on unmount", () => {
    const channel = makeChannel();
    vi.spyOn(supabaseLib.supabase, "channel").mockReturnValue(channel as any);
    const removeChannelSpy = vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { Wrapper } = makeWrapper();

    const { unmount } = renderHook(() => useTasksRealtime(agencyId), { wrapper: Wrapper });
    unmount();

    expect(removeChannelSpy).toHaveBeenCalledWith(channel);
  });

  it("creates a new channel when agencyId changes", () => {
    const channel1 = makeChannel();
    const channel2 = makeChannel();
    const channelSpy = vi.spyOn(supabaseLib.supabase, "channel")
      .mockReturnValueOnce(channel1 as any)
      .mockReturnValueOnce(channel2 as any);
    vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { Wrapper } = makeWrapper();

    const { rerender } = renderHook(
      ({ id }: { id: string }) => useTasksRealtime(id),
      { wrapper: Wrapper, initialProps: { id: agencyId } }
    );

    rerender({ id: otherAgencyId });

    expect(channelSpy).toHaveBeenCalledTimes(2);
    expect(channelSpy).toHaveBeenNthCalledWith(1, `tasks:agency:${agencyId}`);
    expect(channelSpy).toHaveBeenNthCalledWith(2, `tasks:agency:${otherAgencyId}`);
  });

  it("removes the old channel when agencyId changes", () => {
    const channel1 = makeChannel();
    const channel2 = makeChannel();
    vi.spyOn(supabaseLib.supabase, "channel")
      .mockReturnValueOnce(channel1 as any)
      .mockReturnValueOnce(channel2 as any);
    const removeChannelSpy = vi.spyOn(supabaseLib.supabase, "removeChannel").mockReturnValue(undefined as any);
    const { Wrapper } = makeWrapper();

    const { rerender } = renderHook(
      ({ id }: { id: string }) => useTasksRealtime(id),
      { wrapper: Wrapper, initialProps: { id: agencyId } }
    );

    rerender({ id: otherAgencyId });

    expect(removeChannelSpy).toHaveBeenCalledWith(channel1);
  });
});
