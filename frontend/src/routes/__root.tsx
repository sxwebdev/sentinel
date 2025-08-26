import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { Toaster } from "sonner";

import Header from "@/app/layouts/parts/Header";
import NotFound from "@/pages/notFound/notFound";

const RootComponent = () => (
  <>
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 p-6 md:py-8 xl:px-0">
      <Header />
      <Outlet />
    </div>
    <Toaster />
    <TanStackRouterDevtools />
  </>
);

export const Route = createRootRoute({
  component: RootComponent,
  notFoundComponent: NotFound,
});
