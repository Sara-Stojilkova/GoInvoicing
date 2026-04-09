import type { User } from "../types/api";
import { request } from "./client";

export function listUsers(agencyId: string): Promise<User[]> {
  return request<User[]>(`/api/users?agency_id=${agencyId}`);
}

export function getUser(id: string): Promise<User> {
  return request<User>(`/api/users/${id}`);
}

export function createUser(data: { name: string; email: string; role: string; agency_id: string }): Promise<User> {
  return request<User>("/api/users", {
    method: "POST",
    body: JSON.stringify(data),
  });
}
