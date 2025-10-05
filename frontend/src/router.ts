import { createRouter } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";
import { Loader } from "./entities/loader/loader";
import NotFound from "./pages/notFound/notFound";
import ErrorRouter from "./entities/errorRouter";

// Create a new router instance
export const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  defaultStaleTime: Infinity,
  scrollRestoration: true,
  defaultPendingComponent: Loader,
  defaultNotFoundComponent: NotFound,
  defaultErrorComponent: ErrorRouter,
  context: {
    // auth will be passed down from App component
    auth: undefined!,
  },
});

// Register the router instance for type safety
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
