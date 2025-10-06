import type { User } from "@/api/gen/sentinel/users/v1/users_pb";
import { createContext, useContext } from "react";

const AUTH_STORE_NAME = "authSession";

export type AuthSession = {
  accessToken: string;
  refreshToken: string;
  accessTokenExpiredAt: string;
  refreshTokenExpiredAt: string;
  deviceId: string;
};

export type AuthError = {
  name: string;
  message: string;
};

export interface AuthState {
  isLoading: boolean;
  isAuthenticated: boolean;
  user?: User;
  authError?: AuthError;
  authorization: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

export function isAuthSession(v: unknown): v is AuthSession {
  if (typeof v !== "object" || v === null) return false;
  const o = v as Record<string, unknown>;
  return (
    typeof o.accessToken === "string" &&
    typeof o.refreshToken === "string" &&
    typeof o.accessTokenExpiredAt === "string" &&
    typeof o.refreshTokenExpiredAt === "string" &&
    typeof o.deviceId === "string"
  );
}

// clearSession from localStorage
export function clearSession() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(AUTH_STORE_NAME);
}

// saveSession to localStorage
export function saveSession(session: AuthSession) {
  if (typeof window === "undefined") return;
  localStorage.setItem(AUTH_STORE_NAME, JSON.stringify(session));
}

// loadSession from localStorage
export function loadSession(): AuthSession | null {
  if (typeof window === "undefined") return null;
  const session = localStorage.getItem(AUTH_STORE_NAME);
  if (!session) return null;
  try {
    const parsed = JSON.parse(session);
    if (isAuthSession(parsed)) {
      return parsed;
    }
    return null;
  } catch {
    clearSession();
    return null;
  }
}

export const AuthContext = createContext<AuthState | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
