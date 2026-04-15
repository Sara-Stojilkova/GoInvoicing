export interface Agency {
  id: string;
  name: string;
  created_at: string;
}

export interface User {
  id: string;
  name: string;
  email: string;
  role: "admin" | "member";
  agency_id: string;
  created_at: string;
}

export interface Task {
  id: string;
  title: string;
  description: string | null;
  status: "todo" | "in_progress" | "done";
  priority: "low" | "medium" | "high";
  agency_id: string;
  assignee_id: string | null;
  created_at: string;
  due_date: string | null;
  completed_at: string | null;
}
