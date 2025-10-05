import NotificationHistoryPage from "@/pages/settings/notification-history";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/settings/notifications-history")({
  component: RouteComponent,
});

function RouteComponent() {
  return <NotificationHistoryPage />;
}
