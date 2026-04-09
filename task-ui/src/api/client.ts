import { ApiError } from "./error";

export async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  const body = await response.json();

  if (!response.ok) {
    throw new ApiError(response.status, body.error ?? response.statusText);
  }

  return body as T;
}
