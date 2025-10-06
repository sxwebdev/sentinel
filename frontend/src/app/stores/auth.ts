import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User } from "@/api/gen/sentinel/users/v1/users_pb";
import { type AuthorizationResponse } from "@/api/gen/sentinel/auth/v1/auth_pb";
import {
  AuthorizationRequestSchema,
  AuthenticateRequestSchema,
  RefreshTokenRequestSchema,
  LogoutRequestSchema,
  DeviceInfoSchema,
  DeviceType,
} from "@/api/gen/sentinel/auth/v1/auth_pb";
import { Code, ConnectError } from "@connectrpc/connect";
import { create as createMsg } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";

import { authClient } from "@/api/api";

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

export interface AuthStoreState {
  // state
  isLoading: boolean;
  isAuthenticated: boolean;
  user?: User;
  authError?: AuthError;
  session?: AuthSession | null;

  // actions
  init: () => Promise<void>;
  authorization: (email: string, password: string) => Promise<void>;
  authenticate: (accessToken: string) => Promise<User | undefined>;
  refreshToken: () => Promise<void>;
  logout: () => Promise<void>;
  clear: () => void;
}

export const useAuthStore = create<AuthStoreState>()(
  persist(
    (set, get) => ({
      isLoading: true,
      isAuthenticated: false,
      user: undefined,
      authError: undefined,
      session: null,

      clear: () => {
        set({
          isLoading: false,
          isAuthenticated: false,
          user: undefined,
          authError: undefined,
          session: null,
        });
      },

      init: async () => {
        const session = get().session;
        if (!session) {
          set({ isLoading: false, isAuthenticated: false });
          return;
        }

        // Refresh token expired -> clear
        if (Date.now() >= new Date(session.refreshTokenExpiredAt).getTime()) {
          get().clear();
          return;
        }

        let currentSession = session;
        // Access token expired -> try refresh
        if (
          Date.now() >= new Date(currentSession.accessTokenExpiredAt).getTime()
        ) {
          try {
            const res = await authClient.refreshToken(
              createMsg(RefreshTokenRequestSchema, {
                refreshToken: currentSession.refreshToken,
              }),
            );
            if (!res.accessTokenExpiredAt || !res.refreshTokenExpiredAt) {
              throw new Error("missing token expiration in response");
            }
            currentSession = {
              accessToken: res.accessToken,
              refreshToken: res.refreshToken,
              accessTokenExpiredAt: timestampDate(
                res.accessTokenExpiredAt,
              ).toISOString(),
              refreshTokenExpiredAt: timestampDate(
                res.refreshTokenExpiredAt,
              ).toISOString(),
              deviceId: currentSession.deviceId,
            };
            set({ session: currentSession, isAuthenticated: true });
          } catch (e) {
            if (e instanceof ConnectError) {
              set({ authError: { name: e.name, message: e.rawMessage } });
            }
            get().clear();
            return;
          }
        }

        // Authenticate to fetch user
        try {
          const res = await authClient.authenticate(
            createMsg(AuthenticateRequestSchema, {
              accessToken: currentSession.accessToken,
            }),
          );
          set({ user: res.user, isAuthenticated: true, isLoading: false });
        } catch (e) {
          if (e instanceof ConnectError) {
            if (e.code === Code.Unauthenticated) {
              get().clear();
              return;
            }
            set({ authError: { name: e.name, message: e.rawMessage } });
          }
          set({ isLoading: false });
        }
      },

      authenticate: async (accessToken: string) => {
        try {
          const res = await authClient.authenticate(
            createMsg(AuthenticateRequestSchema, { accessToken }),
          );
          set({ user: res.user, isAuthenticated: true });
          return res.user;
        } catch (e) {
          if (e instanceof ConnectError) {
            set({ authError: { name: e.name, message: e.rawMessage } });
          }
          return undefined;
        }
      },

      authorization: async (email: string, password: string) => {
        set({ authError: undefined });
        try {
          const res: AuthorizationResponse = await authClient.authorization(
            createMsg(AuthorizationRequestSchema, {
              email,
              password,
              deviceInfo: createMsg(DeviceInfoSchema, {
                deviceId: get().session?.deviceId || "",
                deviceType: DeviceType.WEB,
                deviceName: navigator.userAgent,
              }),
            }),
          );
          if (
            !res.authPayload ||
            !res.authPayload.accessTokenExpiredAt ||
            !res.authPayload.refreshTokenExpiredAt
          ) {
            throw new Error("missing token expiration in response");
          }
          const session: AuthSession = {
            accessToken: res.authPayload.accessToken,
            refreshToken: res.authPayload.refreshToken,
            accessTokenExpiredAt: timestampDate(
              res.authPayload.accessTokenExpiredAt,
            ).toISOString(),
            refreshTokenExpiredAt: timestampDate(
              res.authPayload.refreshTokenExpiredAt,
            ).toISOString(),
            deviceId: res.authPayload.deviceId,
          };
          set({ session, user: res.user, isAuthenticated: true });
        } catch (e) {
          if (e instanceof ConnectError) {
            set({ authError: { name: e.name, message: e.rawMessage } });
          } else if (e instanceof Error) {
            set({ authError: { name: e.name, message: e.message } });
          }
          throw e;
        }
      },

      refreshToken: async (): Promise<void> => {
        const session = get().session;
        if (!session) {
          get().clear();
          return;
        }

        // If refresh token expired, signal Unauthenticated-like condition, but don't clear here
        if (Date.now() >= new Date(session.refreshTokenExpiredAt).getTime()) {
          get().clear();
          return;
        }

        try {
          const res = await authClient.refreshToken(
            createMsg(RefreshTokenRequestSchema, {
              refreshToken: session.refreshToken,
            }),
          );
          if (!res.accessTokenExpiredAt || !res.refreshTokenExpiredAt) {
            throw new Error("missing token expiration in response");
          }
          const newSession: AuthSession = {
            accessToken: res.accessToken,
            refreshToken: res.refreshToken,
            accessTokenExpiredAt: timestampDate(
              res.accessTokenExpiredAt,
            ).toISOString(),
            refreshTokenExpiredAt: timestampDate(
              res.refreshTokenExpiredAt,
            ).toISOString(),
            deviceId: session.deviceId,
          };
          set({ session: newSession, isAuthenticated: true });
          return;
        } catch (e) {
          if (e instanceof ConnectError && e.code === Code.Unauthenticated) {
            // Only clear on explicit Unauthenticated when called by interceptor per requirement
            get().clear();
          }
        }
      },

      logout: async () => {
        try {
          await authClient.logout(createMsg(LogoutRequestSchema));
        } catch {
          // ignore network/logging errors
        } finally {
          get().clear();
        }
      },
    }),
    {
      name: "authStore",
      version: 1,
      partialize: (state) => ({ session: state.session }),
      onRehydrateStorage: () => (state) => {
        // After hydration, kick off init to validate tokens and fetch user
        if (state) {
          state.init().catch(() => {
            state.clear();
          });
        }
      },
    },
  ),
);
