import type { User } from "@/api/gen/sentinel/users/v1/users_pb";
import { createContext, useContext } from "react";

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
  // refreshToken: () => Promise<RefreshTokenResponse>;
  logout: () => void;
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

export const AuthContext = createContext<AuthState | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
