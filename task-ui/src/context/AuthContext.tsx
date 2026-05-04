import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { loginApi, registerApi } from "../api/auth";
import type { RegisterRequest } from "../api/auth";

interface AuthState {
  token: string | null;
  agencyId: string | null;
  role: string | null;
  userEmail: string | null;
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>;
  register: (req: RegisterRequest) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const KEYS = {
  token: "auth_token",
  agencyId: "auth_agency_id",
  role: "auth_role",
  email: "auth_email",
} as const;

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>(() => ({
    token: localStorage.getItem(KEYS.token),
    agencyId: localStorage.getItem(KEYS.agencyId),
    role: localStorage.getItem(KEYS.role),
    userEmail: localStorage.getItem(KEYS.email),
  }));

  const logout = useCallback(() => {
    Object.values(KEYS).forEach((k) => localStorage.removeItem(k));
    setState({ token: null, agencyId: null, role: null, userEmail: null });
  }, []);

  useEffect(() => {
    const handle = () => logout();
    window.addEventListener("auth:logout", handle);
    return () => window.removeEventListener("auth:logout", handle);
  }, [logout]);

  const login = useCallback(async (email: string, password: string) => {
    const res = await loginApi({ email, password });
    const agencyId = res.user.app_metadata.agency_id;
    const role = res.user.app_metadata.role;
    localStorage.setItem(KEYS.token, res.access_token);
    localStorage.setItem(KEYS.agencyId, agencyId);
    localStorage.setItem(KEYS.role, role);
    localStorage.setItem(KEYS.email, res.user.email);
    setState({ token: res.access_token, agencyId, role, userEmail: res.user.email });
  }, []);

  const register = useCallback(async (req: RegisterRequest) => {
    await registerApi(req);
  }, []);

  return (
    <AuthContext.Provider value={{ ...state, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
