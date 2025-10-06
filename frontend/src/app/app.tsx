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

  const authStore = useAuthStore(
    useShallow((s) => ({
      isLoading: s.isLoading,
      isAuthenticated: s.isAuthenticated,
    })),
  );

  const systemStore = useSystemStore(
    useShallow((s) => ({
      isLoading: s.isLoading,
      isSystemInitialized: s.isSystemInitialized,
    })),
  );

  // Initialize theme on app load
  useEffect(() => {
    initTheme();
  }, [initTheme]);

  if (systemStore.isLoading || authStore.isLoading) {
    return <PageLoader />;
  }

  return (
    <RouterProvider
      router={router}
      context={{
        isAuthenticated: authStore.isAuthenticated,
        isSystemInitialized: systemStore.isSystemInitialized,
      }}
    />
  );
};

export default App;
