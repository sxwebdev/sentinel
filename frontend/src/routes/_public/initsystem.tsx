import { InitSystemPage } from "@/pages/public/initSystem";
import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/initsystem")({
  beforeLoad: ({ context }) => {
    if (context.isSystemInitialized) {
      throw redirect({
        to: "/login",
      });
    }
  },
  component: () => <InitSystemPage />,
});
