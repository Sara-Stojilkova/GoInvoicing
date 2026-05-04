import { ApiError } from "./error";

export async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const token = localStorage.getItem("auth_token");

  const response = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });

  const contentType = response.headers.get("Content-Type") ?? "";
  const body = contentType.includes("application/json") ? await response.json() : undefined;

  if (!response.ok) {
    if (response.status === 401) {
      window.dispatchEvent(new Event("auth:logout"));
    }
    throw new ApiError(response.status, body?.error ?? response.statusText);
  }

  return body as T;
}
