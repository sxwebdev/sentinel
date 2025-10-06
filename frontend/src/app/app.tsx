import { RouterProvider } from "@tanstack/react-router";

import { useAuthStore } from "./stores/auth";
import { router } from "../router";
import { useSystemStore } from "./stores/system";
import PageLoader from "@/shared/components/pageLoader";

const App = () => {
  const authStore = useAuthStore();
  const systemStore = useSystemStore();

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
