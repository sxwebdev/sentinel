import React, { useState, useEffect } from "react";
import {
  AuthContext,
  isAuthSession,
  type AuthError,
  type AuthSession,
} from "./hooks";
import type { User } from "@/api/gen/sentinel/users/v1/users_pb";
import { toast } from "sonner";
import { authClient } from "@/api/api";
import { ConnectError } from "@connectrpc/connect";
import { create } from "@bufbuild/protobuf";
import {
  AuthenticateRequestSchema,
  AuthorizationRequestSchema,
  DeviceInfoSchema,
  DeviceType,
  LogoutRequestSchema,
  RefreshTokenRequestSchema,
} from "@/api/gen/sentinel/auth/v1/auth_pb";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import PageLoader from "@/shared/components/pageLoader";

const AUTH_STORE_NAME = "authSession";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [authError, setAuthError] = useState<AuthError | undefined>(undefined);
  const [user, setUser] = useState<User | undefined>(undefined);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const clearStore = () => {
    localStorage.removeItem(AUTH_STORE_NAME);
    setIsLoading(false);
    setIsAuthenticated(false);
    setUser(undefined);
    setAuthError(undefined);
  };

  // Restore auth state on app load
  useEffect(() => {
    const auth = async () => {
      const sessionStr = localStorage.getItem(AUTH_STORE_NAME);
      if (sessionStr) {
        let session: AuthSession | null = null;

        try {
          const parsed = JSON.parse(sessionStr) as unknown;
          if (isAuthSession(parsed)) session = parsed;
        } catch {
          toast.error("failed to parse stored session");
        }

        if (!session) {
          clearStore();
          return;
        }

        // Check if access token and refresh token are expired
        if (Date.now() >= new Date(session.refreshTokenExpiredAt).getTime()) {
          clearStore();
          return;
        }

        // If access token is expired
        if (Date.now() >= new Date(session.accessTokenExpiredAt).getTime()) {
          try {
            const res = await authClient.refreshToken(
              create(RefreshTokenRequestSchema, {
                refreshToken: session.refreshToken,
              }),
            );

            if (!res.accessTokenExpiredAt || !res.refreshTokenExpiredAt) {
              throw new Error("missing token expiration in response");
            }

            session = {
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

            localStorage.setItem(AUTH_STORE_NAME, JSON.stringify(session));
            setIsAuthenticated(true);
          } catch (error) {
            if (error instanceof ConnectError) {
              setAuthError({
                name: error.name,
                message: error.rawMessage,
              });

              toast.error(`Authentication error: ${error.rawMessage}`);
            } else {
              toast.error(`Unexpected error: ${String(error)}`);
            }
            clearStore();
            return;
          }
        }

        // Authenticate user
        try {
          const res = await authClient.authenticate(
            create(AuthenticateRequestSchema, {
              accessToken: session.accessToken,
            }),
          );

          setUser(res.user);
        } catch (error) {
          if (error instanceof ConnectError) {
            setAuthError({
              name: error.name,
              message: error.rawMessage,
            });

            toast.error(`Authentication error: ${error.rawMessage}`);
          } else {
            toast.error(`Unexpected error: ${String(error)}`);
          }
          return;
        }

        setIsAuthenticated(true);
        setIsLoading(false);
      } else {
        setIsLoading(false);
      }
    };

    void auth();
  }, []);

  const logout = async () => {
    try {
      await authClient.logout(create(LogoutRequestSchema));
    } catch (error) {
      toast.error(
        "Failed to logout" +
          (error instanceof ConnectError ? `: ${error.rawMessage}` : ""),
      );
    } finally {
      clearStore();
    }
  };

  const authorization = async (email: string, password: string) => {
    setAuthError(undefined);

    try {
      // get deviceID from existing session or create a new one
      let deviceId = "";
      const sessionStr = localStorage.getItem(AUTH_STORE_NAME);
      if (sessionStr) {
        const parsed = JSON.parse(sessionStr) as unknown;
        if (isAuthSession(parsed)) deviceId = parsed.deviceId;
      }

      const res = await authClient.authorization(
        create(AuthorizationRequestSchema, {
          email,
          password,
          deviceInfo: create(DeviceInfoSchema, {
            deviceId: deviceId,
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

      localStorage.setItem(AUTH_STORE_NAME, JSON.stringify(session));

      setUser(res.user);
      setIsAuthenticated(true);
    } catch (error) {
      if (error instanceof ConnectError) {
        setAuthError({
          name: error.name,
          message: error.rawMessage,
        });

        toast.error(`Authentication error: ${error.rawMessage}`);
      } else {
        toast.error(`Unexpected error: ${String(error)}`);
      }
      throw error;
    }
  };

  useEffect(() => {
    const timer = setInterval(async () => {
      const sessionStr = localStorage.getItem(AUTH_STORE_NAME);
      if (!sessionStr) {
        clearStore();
        return;
      }

      let session: AuthSession | null = null;

      try {
        const parsed = JSON.parse(sessionStr) as unknown;
        if (isAuthSession(parsed)) session = parsed;
      } catch {
        toast.error("failed to parse stored session");
      }

      if (!session) {
        clearStore();
        return;
      }

      // Check if refresh token is expired
      if (Date.now() >= new Date(session.refreshTokenExpiredAt).getTime()) {
        clearStore();
        return;
      }

      // If access token is valid for more than 2 minutes, skip refresh
      const msLeft =
        new Date(session.accessTokenExpiredAt).getTime() - Date.now();
      if (msLeft > 2 * 60 * 1000) {
        return; // still far from expiry
      }

      try {
        const res = await authClient.refreshToken(
          create(RefreshTokenRequestSchema, {
            refreshToken: session.refreshToken,
          }),
        );

        if (!res.accessTokenExpiredAt || !res.refreshTokenExpiredAt) {
          throw new Error("missing token expiration in response");
        }

        session = {
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

        localStorage.setItem(AUTH_STORE_NAME, JSON.stringify(session));
        setIsAuthenticated(true);
      } catch (error) {
        if (error instanceof ConnectError) {
          setAuthError({
            name: error.name,
            message: error.rawMessage,
          });

          toast.error(`Authentication error: ${error.rawMessage}`);
        } else {
          toast.error(`Unexpected error: ${String(error)}`);
        }
      }
    }, 60 * 1000); // every 1 minute
    return () => clearInterval(timer);
  }, [setIsAuthenticated]);

  // Show loading state while checking auth
  if (isLoading) {
    return <PageLoader />;
  }

  return (
    <AuthContext.Provider
      value={{
        isLoading,
        isAuthenticated,
        authError,
        user,
        authorization,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
