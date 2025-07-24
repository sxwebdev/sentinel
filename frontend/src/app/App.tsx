import {ErrorBoundary} from "@/shared/components/ErrorBoundary";
import {RouterProvider} from "react-router";
import {router} from "./routes/routes";
import {Suspense} from "react";
import {Loader} from "@/entities/loader/loader";

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<Loader loaderPage />}>
        <RouterProvider router={router} />
      </Suspense>
    </ErrorBoundary>
  );
}

export default App;
