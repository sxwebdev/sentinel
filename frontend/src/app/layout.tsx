import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/shared/components/ui/sidebar";
import { AppSidebar } from "@/shared/components/sidebar/sidebar";
import { ThemeToggle } from "@/shared/components/theme-toggle";
import { SystemInfo } from "@/features/apiInfo/systemInfo";

import type { Project } from "@/api/gen/sentinel/projects/v1/projects_pb";

type LayoutProps = {
  children: React.ReactNode;
  projects: Project[];
};

export default function Layout({ children, projects }: LayoutProps) {
  return (
    <SidebarProvider>
      <AppSidebar projects={projects} />
      <SidebarInset>
        <header className="sticky top-0 z-10 flex h-16 shrink-0 items-center justify-between gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
          </div>
          <div className="flex items-center gap-2 px-4">
            <SystemInfo />
            <ThemeToggle />
          </div>
        </header>
        <div className="bg-background text-foreground flex-1 overflow-y-auto p-4 pt-0">
          <div className="container mx-auto max-w-6xl py-6">{children}</div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
