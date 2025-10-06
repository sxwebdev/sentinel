import { createRoot } from "react-dom/client";
import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";

import "@shared/styles/index.css";

import { connectQueryClient, connectTransport } from "./api/api";
import App from "@app/app";

const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = createRoot(rootElement);
  root.render(
    <TransportProvider transport={connectTransport}>
      <QueryClientProvider client={connectQueryClient}>
        <App />
      </QueryClientProvider>
    </TransportProvider>,
  );
}
