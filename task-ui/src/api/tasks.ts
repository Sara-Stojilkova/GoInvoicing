import type { Task } from "../types/api";
import { request } from "./client";

export function listTasks(agencyId: string): Promise<Task[]> {
  return request<Task[]>(`/api/tasks?agency_id=${agencyId}`);
}

export function getTask(id: string, agencyId: string): Promise<Task> {
  return request<Task>(`/api/tasks/${id}?agency_id=${agencyId}`);
}

export function createTask(data: {
  title: string;
  priority: string;
  agency_id: string;
  description?: string;
  assignee_id?: string;
  due_date?: string;
  tags?: string[];
}): Promise<Task> {
  const cleaned = Object.fromEntries(
    Object.entries(data).filter(([ , v]) => v != undefined)
  );
  if (cleaned.due_date) {
    cleaned.due_date = new Date(cleaned.due_date as string).toISOString();
  }
  return request<Task>("/api/tasks", {
    method: "POST",
    body: JSON.stringify(cleaned),
  });
}

export function assignTask(id: string, data: { assignee_id: string; assignee_agency_id: string }): Promise<void> {
  return request<void>(`/api/tasks/${id}/assign`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function unassignTask(id: string): Promise<void> {
  return request<void>(`/api/tasks/${id}/unassign`, { method: "POST" });
}

export function completeTask(id: string): Promise<void> {
  return request<void>(`/api/tasks/${id}/complete`, { method: "POST" });
}

export function setTaskInProgress(id: string): Promise<void> {
  return request<void>(`/api/tasks/${id}/set-in-progress`, { method: "POST" });
}

export function updateDescription(id: string, description: string | null): Promise<void> {
  return request<void>(`/api/tasks/${id}/description`, {
    method: "PATCH",
    body: JSON.stringify({ description }),
  });
}

export function updateDueDate(id: string, dueDate: string | null): Promise<void> {
  return request<void>(`/api/tasks/${id}/due-date`, {
    method: "PATCH",
    body: JSON.stringify({ due_date: dueDate }),
  });
}

export function updateTags(id: string, tags: string[]): Promise<void> {
  return request<void>(`/api/tasks/${id}/tags`, {
    method: "PATCH",
    body: JSON.stringify({ tags }),
  });
}
