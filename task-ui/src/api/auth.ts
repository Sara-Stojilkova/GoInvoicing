import { request } from "./client";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: {
    email: string;
    app_metadata: {
      agency_id: string;
      role: string;
    };
  };
}

export interface RegisterRequest {
  full_name: string;
  email: string;
  password: string;
  agency_name?: string;
  agency_id?: string;
}

export interface RegisterResponse {
  user_id: string;
  agency_id: string;
  role: string;
  activated: boolean;
}

export function loginApi(req: LoginRequest): Promise<LoginResponse> {
  return request<LoginResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function registerApi(req: RegisterRequest): Promise<RegisterResponse> {
  return request<RegisterResponse>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify(req),
  });
}
