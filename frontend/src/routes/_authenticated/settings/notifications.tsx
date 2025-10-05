import { createFileRoute } from "@tanstack/react-router";
import NotificationsPage from "@/pages/settings/notifications";

export const Route = createFileRoute("/_authenticated/settings/notifications")({
  component: RouteComponent,
});

function RouteComponent() {
  return <NotificationsPage />;
}
