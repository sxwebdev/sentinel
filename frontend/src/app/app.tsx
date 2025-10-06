import { RouterProvider } from "@tanstack/react-router";

import { useAuth } from "./providers/auth/context";
import { AuthProvider } from "./providers/auth";
import { router } from "../router";

function InnerApp() {
  const auth = useAuth();
  return <RouterProvider router={router} context={{ auth }} />;
}

const App = () => {
  return (
    <AuthProvider>
      <InnerApp />
    </AuthProvider>
  );
};

export default App;
