import { ChevronsUpDown, Plus } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/shared/components/ui/sidebar";
import { useProjectStore } from "@/app/stores/projects";
import { useShallow } from "zustand/react/shallow";
import {
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Input,
} from "../ui";
import { Field, FieldGroup, FieldLabel } from "../ui/field";
import { useState } from "react";
import { toast } from "sonner";
import { Spinner } from "../ui/spinner";
import { cn } from "@/shared/lib/utils";
import LogoSmall from "../logoSmall";

type AddProjectProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
};

const AddProject = ({ open, setOpen }: AddProjectProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const projectsStore = useProjectStore();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    e.stopPropagation();

    if (!name) {
      toast.error("Please fill in all required fields");
      return;
    }

    try {
      setIsLoading(true);
      await projectsStore.createProject(name, description);
      setOpen(false);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild></DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Project</DialogTitle>
          <DialogDescription>
            Fill out the form below to create a new project.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-8">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="name" required>
                Name
              </FieldLabel>
              <Input
                id="name"
                type="text"
                placeholder="Name"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="description">Description</FieldLabel>
              <Input
                id="description"
                type="text"
                placeholder="Description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button disabled={isLoading}>
              {isLoading ? <Spinner /> : "Create Project"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export function ProjectSwitcher() {
  const [open, setOpen] = useState(false);
  const { isMobile } = useSidebar();
  const { projects, selectedProjectId, selectedProject, selectProject } =
    useProjectStore(
      useShallow((state) => ({
        projects: state.projects,
        selectedProjectId: state.selectedProjectId,
        selectedProject: state.selectedProject,
        selectProject: state.selectProject,
      })),
    );

  if (!selectedProjectId) {
    return null;
  }

  const activeProject = selectedProject();

  return (
    <>
      <AddProject open={open} setOpen={setOpen} />
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <LogoSmall className="size-4 fill-white transition-all dark:fill-white/60" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">
                    {activeProject?.name}
                  </span>
                  <span className="truncate text-xs">
                    {activeProject?.description}
                  </span>
                </div>
                <ChevronsUpDown className="ml-auto" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
              align="start"
              side={isMobile ? "bottom" : "right"}
              sideOffset={4}
            >
              <DropdownMenuLabel className="text-muted-foreground text-xs">
                Projects
              </DropdownMenuLabel>
              {projects.map((project) => (
                <DropdownMenuItem
                  key={project.id}
                  onClick={() => selectProject(project.id)}
                  className={cn("cursor-pointer gap-2 p-2", {
                    "bg-sidebar-accent text-sidebar-accent-foreground":
                      project.id === selectedProjectId,
                  })}
                >
                  {project.name}
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />

              <DropdownMenuItem
                className="cursor-pointer gap-2 p-2"
                onClick={() => setOpen(true)}
              >
                <div className="flex size-6 items-center justify-center rounded-md border bg-transparent">
                  <Plus className="size-4" />
                </div>
                <div className="text-muted-foreground font-medium">
                  Add project
                </div>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </>
  );
}
