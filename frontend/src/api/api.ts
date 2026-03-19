import {
  Code,
  ConnectError,
  createClient,
  type Interceptor,
} from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryClient } from "@tanstack/react-query";

import { NotificationService } from "./gen/sentinel/notifications/v1/notifications_pb";
import { AgentsService } from "./gen/sentinel/agents/v1/agents_pb";
import { AuthService } from "./gen/sentinel/auth/v1/auth_pb";
import { ProjectsService } from "./gen/sentinel/projects/v1/projects_pb";
import { SystemService } from "./gen/sentinel/system/v1/service_pb";
import { useAuthStore } from "@/app/stores/auth";
import { useProjectStore } from "@/app/stores/projects";

export const VITE_SERVER_API_BASE_URL =
  import.meta.env.VITE_SERVER_API_BASE_URL || window.location.origin + "/api";

const skipAuthMethods = ["RefreshToken", "Authorization"];

const authInterceptor: Interceptor = (next) => async (req) => {
  const { session } = useAuthStore.getState();
  const { selectedProjectId } = useProjectStore.getState();

  if (selectedProjectId) {
    req.header.set("X-Project-ID", selectedProjectId);
  }

  if (session?.accessToken) {
    req.header.set("Authorization", `Bearer ${session.accessToken}`);
  }

  try {
    return await next(req);
  } catch (error) {
    if (
      error instanceof ConnectError &&
      error.code === Code.Unauthenticated &&
      !skipAuthMethods.includes(req.method.name)
    ) {
      try {
        await useAuthStore.getState().refreshToken();
        const freshSession = useAuthStore.getState().session;
        if (!freshSession?.accessToken) {
          throw new Error("No access token");
        }

        req.header.set(
          "Authorization",
          `Bearer ${freshSession.accessToken}`,
        );
        return await next(req);
      } catch (refreshError) {
        console.error("Failed to refresh token:", refreshError);
        throw refreshError;
      }
    }
    throw error;
  }
};

export const connectTransport = createConnectTransport({
  baseUrl: VITE_SERVER_API_BASE_URL,
  interceptors: [authInterceptor],
});

export const connectQueryClient = new QueryClient();

export const systemClient = createClient(SystemService, connectTransport);

export const authClient = createClient(AuthService, connectTransport);

export const projectsClient = createClient(ProjectsService, connectTransport);

export const agentsClient = createClient(AgentsService, connectTransport);

export const notificationsClient = createClient(
  NotificationService,
  connectTransport,
);
