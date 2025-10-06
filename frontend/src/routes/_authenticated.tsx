import ProtectedWrapper from "@/app/protected";
import { createFileRoute, redirect, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: ({ context }) => {
    if (!context.isSystemInitialized) {
      throw redirect({
        to: "/initsystem",
      });
    }
    if (!context.isAuthenticated) {
      throw redirect({
        to: "/login",
      });
    }
  },
  component: () => (
    <ProtectedWrapper>
      <Outlet />
    </ProtectedWrapper>
  ),
});
