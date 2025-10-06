import { RouterProvider } from "@tanstack/react-router";

import { useAuthStore } from "./stores/auth";
import { router } from "../router";
import { useSystemStore } from "./stores/system";
import PageLoader from "@/shared/components/pageLoader";
import { useUserPreferenceStore } from "./stores/userPreferience";
import { useEffect } from "react";
import { useShallow } from "zustand/react/shallow";

const App = () => {
  const { initTheme } = useUserPreferenceStore(
    useShallow((s) => ({ initTheme: s.initTheme })),
  );

  const {
    isLoading: isSystemLoading,
    isSystemInitialized,
    checkIsInitialized,
    error,
  } = useSystemStore(
    useShallow((s) => ({
      isLoading: s.isLoading,
      error: s.error,
      isSystemInitialized: s.isSystemInitialized,
      checkIsInitialized: s.checkIsInitialized,
    })),
  );

  const { isLoading, isAuthenticated, init } = useAuthStore(
    useShallow((s) => ({
      isLoading: s.isLoading,
      isAuthenticated: s.isAuthenticated,
      init: s.init,
    })),
  );

  // Initialize theme on app load
  useEffect(() => {
    initTheme();
  }, [initTheme]);

  // Initialize system state on app load
  useEffect(() => {
    checkIsInitialized();
  }, [checkIsInitialized]);

  // Initialize auth on app load
  useEffect(() => {
    if (!isSystemLoading) return;
    init();
  }, [isSystemLoading, init]);

  if (error) {
    throw error;
  }

  if (isSystemLoading || isLoading) {
    return <PageLoader />;
  }

  return (
    <RouterProvider
      router={router}
      context={{
        isAuthenticated: isAuthenticated,
        isSystemInitialized: isSystemInitialized,
      }}
    />
  );
};

export default App;
