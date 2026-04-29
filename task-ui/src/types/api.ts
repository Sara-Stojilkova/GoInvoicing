import type { Database } from "../database.types";

type Tables<T extends keyof Database["public"]["Tables"]> =
  Database["public"]["Tables"][T]["Row"];

export type Agency = Tables<"agencies">;
export type Task = Tables<"tasks">;

// public.users holds profile data only; email and role live in auth.users / JWT.
export type User = Tables<"users">;

// Convenience re-exports for the enum types
export type TaskStatus = Database["public"]["Enums"]["task_status"];
export type TaskPriority = Database["public"]["Enums"]["task_priority"];

// Utility types for insert/update operations
export type TaskInsert = Database["public"]["Tables"]["tasks"]["Insert"];
export type TaskUpdate = Database["public"]["Tables"]["tasks"]["Update"];
