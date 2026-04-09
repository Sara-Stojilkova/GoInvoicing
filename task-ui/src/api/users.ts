import type { User } from "../types/api";

export function listUsers(_agencyId: string): Promise<User[]> {
  throw new Error("not implemented");
}

export function getUser(_id: string): Promise<User> {
  throw new Error("not implemented");
}

export function createUser(_data: { name: string; email: string; role: string; agency_id: string }): Promise<User> {
  throw new Error("not implemented");
}
