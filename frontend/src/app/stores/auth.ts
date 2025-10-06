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
import { toast } from "sonner";

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
  session?: AuthSession | null;

  // actions
  init: () => Promise<void>;
  authorization: (email: string, password: string) => Promise<void>;
  refreshToken: () => Promise<void>;
  logout: () => Promise<void>;
  clear: () => void;
}

export const useAuthStore = create<AuthStoreState>()(
  persist(
    (set, get) => {
      // Interval for token refresh: checks every 30 seconds
      let refreshInterval: ReturnType<typeof setInterval> | null = null;

      const startRefreshInterval = () => {
        if (refreshInterval) return;

        refreshInterval = setInterval(() => {
          const session = get().session;
          if (!session || !session.accessTokenExpiredAt) return;

          const expiresAt = new Date(session.accessTokenExpiredAt).getTime();
          const now = Date.now();
          const timeLeft = expiresAt - now;

          // If less than 5 minutes left, refresh token
          if (timeLeft < 5 * 60 * 1000) {
            get().refreshToken();
          }
        }, 30 * 1000);
      };

      // Start interval immediately
      startRefreshInterval();

      return {
        isLoading: false,
        isAuthenticated: false,
        user: undefined,
        session: null,

        clear: () => {
          // Stop refresh interval
          if (refreshInterval) {
            clearInterval(refreshInterval);
            refreshInterval = null;
          }

          set({
            isLoading: false,
            isAuthenticated: false,
            user: undefined,
            session: null,
          });
        },

        init: async () => {
          set({ isLoading: true });

          const session = get().session;
          if (!session) {
            get().clear();
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
            Date.now() >=
            new Date(currentSession.accessTokenExpiredAt).getTime()
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
                if (e.code === Code.Unauthenticated) {
                  get().clear();
                }
                toast.error(e.rawMessage);
              }
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
              toast.error(e.rawMessage);
            } else {
              toast.error(
                `Failed to authenticate user ${(e as Error)?.message}`,
              );
            }
            set({ isLoading: false });
          }
        },

        authorization: async (email: string, password: string) => {
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

            // Restart refresh interval after successful authorization
            startRefreshInterval();
          } catch (e) {
            if (e instanceof ConnectError) {
              toast.error(e.rawMessage);
            } else if (e instanceof Error) {
              toast.error(e.message);
            } else {
              toast.error("An unknown error occurred during authorization");
            }
            throw e;
          }
        },

        refreshToken: async (): Promise<void> => {
          const session = get().session;
          if (
            !session ||
            !session.refreshToken ||
            !session.refreshTokenExpiredAt
          ) {
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
            console.log("refresh token error", e);
            get().clear();
            window.location.href = "/login";
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
      };
    },
    {
      name: "authStore",
      version: 1,
      partialize: (state: AuthStoreState) => ({ session: state.session }),
    },
  ),
);
