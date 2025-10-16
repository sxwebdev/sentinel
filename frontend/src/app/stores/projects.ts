import { projectsClient } from "@/api/api";
import {
  ProjectCreateRequestSchema,
  type Project,
} from "@/api/gen/sentinel/projects/v1/projects_pb";
import { ConnectError } from "@connectrpc/connect";
import { toast } from "sonner";
import { create } from "zustand";
import { persist } from "zustand/middleware";
import { create as createProto } from "@bufbuild/protobuf";

type ProjectStore = {
  isLoading: boolean;
  selectedProjectId?: string;
  projects: Project[];
  selectProject: (id: string) => void;
  selectedProject: () => Project | undefined;
  loadProjects: () => Promise<void>;
  createProject: (name: string, description: string) => Promise<void>;
};

export const useProjectStore = create<ProjectStore>()(
  persist(
    (set, get) => ({
      isLoading: true,
      projects: [],
      selectProject: (id: string) => {
        set({ selectedProjectId: id });
        window.location.reload();
      },
      selectedProject: () => {
        const { selectedProjectId, projects } = get();
        const selectedProject = projects.find(
          (p) => p.id === selectedProjectId,
        );

        if (selectedProject) {
          return selectedProject;
        }

        if (projects.length > 0) {
          return projects[0];
        }

        return undefined;
      },
      loadProjects: async () => {
        try {
          const data = await projectsClient.projectsList({});
          set({ projects: data.items });

          const selectedProjectId = get().selectedProjectId;
          if (!selectedProjectId && data.items.length > 0) {
            set({ selectedProjectId: data.items[0].id });
          }
        } catch (error) {
          if (error instanceof ConnectError) {
            toast.error(error.rawMessage);
          } else {
            console.error("Failed to load projects:", error);
          }
        } finally {
          set({ isLoading: false });
        }
      },
      createProject: async (name: string, description: string) => {
        try {
          const res = await projectsClient.projectCreate(
            createProto(ProjectCreateRequestSchema, {
              name,
              description,
            }),
          );

          await get().loadProjects();

          set({ selectedProjectId: res.item?.id });

          window.location.reload();
        } catch (error) {
          if (error instanceof ConnectError) {
            toast.error(error.rawMessage);
          } else {
            console.error("Failed to create project:", error);
          }
        }
      },
    }),
    {
      name: "projectStore",
      partialize: (state) => ({
        selectedProjectId: state.selectedProjectId,
      }),
    },
  ),
);
