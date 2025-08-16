import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { Toaster } from "sonner";

import Header from "@/app/layouts/parts/Header";
import NotFound from "@/pages/notFound/notFound";

const RootComponent = () => (
  <>
    <div className="flex flex-col p-6 md:py-8 xl:px-0 w-full max-w-6xl mx-auto gap-8">
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
