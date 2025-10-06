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

export const VITE_SERVER_API_BASE_URL =
  import.meta.env.VITE_SERVER_API_BASE_URL || window.location.origin + "/api";

const authInterceptor: Interceptor = (next) => async (req) => {
  const { session, clear } = useAuthStore.getState();

  if (session?.accessToken) {
    req.header.set("Authorization", `Bearer ${session.accessToken}`);
  }

  // try {
  //   return await next(req);
  // } catch (error) {
  //   if (error instanceof ConnectError && error.code === Code.Unauthenticated) {
  //     try {
  //       await refreshToken();
  //       if (!session?.accessToken) {
  //         throw new Error("No access token");
  //       }

  //       req.header.set("Authorization", `Bearer ${session?.accessToken}`);
  //       return await next(req);
  //     } catch (refreshError) {
  //       console.error("Failed to refresh token:", refreshError);
  //       clear();
  //       throw refreshError;
  //     }
  //   }
  //   throw error;
  // }

  try {
    return await next(req);
  } catch (error) {
    if (error instanceof ConnectError && error.code === Code.Unauthenticated) {
      clear();
      window.location.href = "/login";
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
