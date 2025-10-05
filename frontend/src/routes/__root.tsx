import { createRootRoute } from "@tanstack/react-router";
import { Toaster } from "sonner";
// import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

import App from "@/app/app";

const RootComponent = () => (
  <>
    <App />
    <Toaster />
    {/* <TanStackRouterDevtools /> */}
  </>
);

export const Route = createRootRoute({
  component: RootComponent,
});
