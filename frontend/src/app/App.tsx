import { ErrorBoundary } from "@/shared/components/ErrorBoundary";
import { RouterProvider } from "react-router";
import { router } from "./routes/routes";
import { Suspense } from "react";
import { Loader } from "@/entities/loader/loader";
import { Toaster } from "sonner";

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<Loader loaderPage />}>
        <RouterProvider router={router} />
      </Suspense>
      <Toaster />
    </ErrorBoundary>
  );
}

export default App;
