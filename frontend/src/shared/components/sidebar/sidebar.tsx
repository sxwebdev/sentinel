import { linkOptions } from "@tanstack/react-router";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/shared/components/ui/sidebar";
import { HouseIcon, SettingsIcon, GlassesIcon } from "lucide-react";

import { ProjectSwitcher } from "./project-switcher";
import { NavUser } from "./nav-user";
import { NavMain } from "./nav-main";
import { UpdateBanner } from "@/features/apiInfo/updateBanner";
import type { Project } from "@/api/gen/sentinel/projects/v1/projects_pb";

const navMain = [
  {
    main: linkOptions({
      to: "/",
      title: "Dashboard",
      icon: HouseIcon,
    }),
  },
  {
    main: linkOptions({
      to: "/agents",
      title: "Agents",
      icon: GlassesIcon,
    }),
  },
  {
    main: linkOptions({
      to: "/settings",
      title: "Settings",
      icon: SettingsIcon,
    }),
    items: [
      linkOptions({
        to: "/settings/notifications",
        title: "Notifications",
      }),
      linkOptions({
        to: "/settings/notifications-history",
        title: "Notifications history",
      }),
    ],
  },
];

const user = {
  name: "admin",
  email: "admin@google.com",
  avatar: "https://avatars.githubusercontent.com/u/124599",
};

type AppSidebarProps = React.ComponentProps<typeof Sidebar> & {
  projects: Project[];
};

export function AppSidebar({ projects, ...props }: AppSidebarProps) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <ProjectSwitcher projects={projects} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMain} />
      </SidebarContent>
      <SidebarFooter>
        <UpdateBanner />
        <NavUser user={user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
