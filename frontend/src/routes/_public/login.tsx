import { LoginPage } from "@/pages/public/loginPage";
import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/login")({
  beforeLoad: ({ context }) => {
    if (!context.isSystemInitialized) {
      throw redirect({
        to: "/initsystem",
      });
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  return <LoginPage />;
}
