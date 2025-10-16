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
      <header className="absolute top-0 z-10 flex h-16 w-full shrink-0 items-center justify-end gap-2 border-b border-b-transparent px-4 transition-[width,height] ease-linear">
        <ThemeToggle />
      </header>
      <div className="w-full max-w-sm">
        <Outlet />
      </div>
    </div>
  );
}
