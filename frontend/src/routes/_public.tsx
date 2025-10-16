import { ThemeToggle } from "@/shared/components/theme-toggle";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_public")({
  beforeLoad: async ({ context }) => {
    if (context.isAuthenticated) {
      throw redirect({ to: "/" });
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="absolute top-4 right-4 md:top-5 md:right-5">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-sm">
        <Outlet />
      </div>
    </div>
  );
}
