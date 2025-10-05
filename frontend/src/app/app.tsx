import { Outlet } from "@tanstack/react-router";
import Layout from "./layout";
import { useQuery } from "@connectrpc/connect-query";
import { projectsList } from "@/api/gen/sentinel/projects/v1/projects-ProjectsService_connectquery";
import { Spinner } from "@/shared/components/ui/spinner";

const App = () => {
  const queryProject = useQuery(projectsList);

  if (queryProject.isLoading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Spinner className="size-6" />
      </div>
    );
  }

  // if (queryProject.data?.items.length === 0) {
  //   return <ProjectCreate />;
  // }

  return (
    <Layout projects={queryProject.data?.items || []}>
      <Outlet />
    </Layout>
  );
};

export default App;
