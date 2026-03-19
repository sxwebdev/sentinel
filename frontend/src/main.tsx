import { createRoot } from "react-dom/client";
import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";

import "@shared/styles/index.css";

import { connectQueryClient, connectTransport } from "./api/api";
import App from "@app/app";
import { ErrorBoundary } from "./entities/errorRouter/ErrorBoundary";

const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = createRoot(rootElement);
  root.render(
    <ErrorBoundary>
      <TransportProvider transport={connectTransport}>
        <QueryClientProvider client={connectQueryClient}>
          <App />
          <Toaster />
        </QueryClientProvider>
      </TransportProvider>
    </ErrorBoundary>,
  );
}
