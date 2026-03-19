import { createFileRoute } from "@tanstack/react-router";
import SettingsLayout from "@/pages/settings/layout";

export const Route = createFileRoute("/_authenticated/settings")({
  component: RouteComponent,
});

function RouteComponent() {
  return <SettingsLayout />;
}
