import PageLoader from "@/shared/components/pageLoader";
import Layout from "./layout";
import { useProjectStore } from "./stores/projects";
import ProjectCreate from "@/pages/projects/create";

type ProtectedWrapperProps = {
  children: React.ReactNode;
};

const ProtectedWrapper = ({ children }: ProtectedWrapperProps) => {
  const projectStore = useProjectStore();

  if (projectStore.isLoading) {
    return <PageLoader />;
  }

  if (projectStore.projects.length === 0) {
    return <ProjectCreate />;
  }

  return <Layout>{children}</Layout>;
};

export default ProtectedWrapper;
