import { Loader } from "@/entities/loader/loader";
import { ErrorBoundary } from "@/shared/components/ErrorBoundary";
import { lazy, Suspense } from "react";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";

function App() {
  const Dashboard = lazy(() => import("@pages/dashboard/dashboard"));
  const ServiceDetail = lazy(() => import("@/pages/service/serviceDetail"));

  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Suspense fallback={<Loader loaderPage />}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/service/:id" element={<ServiceDetail />} />
          </Routes>
        </Suspense>
        <Toaster closeButton />
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
