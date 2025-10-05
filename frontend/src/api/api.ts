import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryClient } from "@tanstack/react-query";

import { NotificationService } from "./gen/sentinel/notifications/v1/notifications_pb";
import { AgentsService } from "./gen/sentinel/agents/v1/agents_pb";
import { AuthService } from "./gen/sentinel/auth/v1/auth_pb";

export const VITE_SERVER_API_BASE_URL =
  import.meta.env.VITE_SERVER_API_BASE_URL || window.location.origin + "/api";

export const connectTransport = createConnectTransport({
  baseUrl: VITE_SERVER_API_BASE_URL,
});

export const connectQueryClient = new QueryClient();

export const authClient = createClient(AuthService, connectTransport);

export const agentsClient = createClient(AgentsService, connectTransport);

export const notificationsClient = createClient(
  NotificationService,
  connectTransport,
);
