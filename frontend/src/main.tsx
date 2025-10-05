import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";

//import App from "@app/App";

import "@shared/styles/index.css";

// Import the generated route tree
import { routeTree } from "./routeTree.gen";
import { Loader } from "@/entities/loader/loader";
import NotFound from "@/pages/notFound/notFound";
import ErrorRouter from "./entities/errorRouter";
import { ThemeProvider } from "./shared/components/theme-provider";
import { connectQueryClient, connectTransport } from "./api/api";

// Create a new router instance
const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  defaultStaleTime: Infinity,
  scrollRestoration: true,
  defaultPendingComponent: Loader,
  defaultNotFoundComponent: NotFound,
  defaultErrorComponent: ErrorRouter,
});

// Register the router instance for type safety
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = createRoot(rootElement);
  root.render(
    <StrictMode>
      <ThemeProvider>
        <TransportProvider transport={connectTransport}>
          <QueryClientProvider client={connectQueryClient}>
            <RouterProvider router={router} />
          </QueryClientProvider>
        </TransportProvider>
      </ThemeProvider>
    </StrictMode>,
  );
}
