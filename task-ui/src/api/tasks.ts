import type { Task } from "../types/api";
import { request } from "./client";

export function listTasks(agencyId: string): Promise<Task[]> {
  return request<Task[]>(`/api/tasks?agency_id=${agencyId}`);
}

export function getTask(id: string, agencyId: string): Promise<Task> {
  return request<Task>(`/api/tasks/${id}?agency_id=${agencyId}`);
}

export function createTask(data: { title: string; priority: string; agency_id: string }): Promise<Task> {
  return request<Task>("/api/tasks", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function assignTask(id: string, data: { assignee_id: string; assignee_agency_id: string }): Promise<void> {
  return request<void>(`/api/tasks/${id}/assign`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function completeTask(id: string): Promise<void> {
  return request<void>(`/api/tasks/${id}/complete`, { method: "POST" });
}

export function setTaskInProgress(id: string): Promise<void> {
  return request<void>(`/api/tasks/${id}/set-in-progress`, { method: "POST" });
}
