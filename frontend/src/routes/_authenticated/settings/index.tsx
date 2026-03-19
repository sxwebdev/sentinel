import { createFileRoute } from "@tanstack/react-router";
import GeneralSettingsPage from "@/pages/settings/general";

export const Route = createFileRoute("/_authenticated/settings/")({
  component: () => <GeneralSettingsPage />,
});
