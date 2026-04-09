import type { Agency } from "../types/api";

export function listAgencies(): Promise<Agency[]> {
  throw new Error("not implemented");
}

export function getAgency(_id: string): Promise<Agency> {
  throw new Error("not implemented");
}

export function createAgency(_name: string): Promise<Agency> {
  throw new Error("not implemented");
}
