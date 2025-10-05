import { Outlet, RouterProvider } from "@tanstack/react-router";
import Layout from "./layout";
import { useQuery } from "@connectrpc/connect-query";
import { projectsList } from "@/api/gen/sentinel/projects/v1/projects-ProjectsService_connectquery";
import PageLoader from "@/shared/components/pageLoader";
import { useAuth } from "./providers/auth/hooks";
import { AuthProvider } from "./providers/auth";
import { router } from "../router";

function InnerApp() {
  const auth = useAuth();
  return <RouterProvider router={router} context={{ auth }} />;
}

const App = () => {
  const queryProject = useQuery(projectsList);

  if (queryProject.isLoading) {
    return <PageLoader />;
  }

  // if (queryProject.data?.items.length === 0) {
  //   return <ProjectCreate />;
  // }

  return (
    <AuthProvider>
      <InnerApp />
    </AuthProvider>
  );

  // return (
  //   <Layout projects={queryProject.data?.items || []}>
  //     <Outlet />
  //   </Layout>
  // );
};

export default App;
