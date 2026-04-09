import type { Agency } from "../types/api";
import { request } from "./client";

export function listAgencies(): Promise<Agency[]> {
  return request<Agency[]>("/api/agencies");
}

export function getAgency(id: string): Promise<Agency> {
  return request<Agency>(`/api/agencies/${id}`);
}

export function createAgency(name: string): Promise<Agency> {
  return request<Agency>("/api/agencies", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}
