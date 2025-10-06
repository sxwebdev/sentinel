import PageLoader from "@/shared/components/pageLoader";
import Layout from "./layout";
import { useProjectStore } from "./stores/projects";
import ProjectCreate from "@/pages/projects/create";
import { useShallow } from "zustand/react/shallow";
import { useEffect } from "react";

type ProtectedWrapperProps = {
  children: React.ReactNode;
};

const ProtectedWrapper = ({ children }: ProtectedWrapperProps) => {
  const { isLoading, projects, loadProjects } = useProjectStore(
    useShallow((s) => ({
      isLoading: s.isLoading,
      projects: s.projects,
      loadProjects: s.loadProjects,
    })),
  );

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  if (isLoading) {
    return <PageLoader />;
  }

  if (projects.length === 0) {
    return <ProjectCreate />;
  }

  return <Layout>{children}</Layout>;
};

export default ProtectedWrapper;
