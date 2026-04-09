import type { Task } from "../types/api";

export function listTasks(_agencyId: string): Promise<Task[]> {
  throw new Error("not implemented");
}

export function getTask(_id: string, _agencyId: string): Promise<Task> {
  throw new Error("not implemented");
}

export function createTask(_data: { title: string; priority: string; agency_id: string }): Promise<Task> {
  throw new Error("not implemented");
}

export function assignTask(_id: string, _data: { assignee_id: string; assignee_agency_id: string }): Promise<void> {
  throw new Error("not implemented");
}

export function completeTask(_id: string): Promise<void> {
  throw new Error("not implemented");
}

export function setTaskInProgress(_id: string): Promise<void> {
  throw new Error("not implemented");
}
