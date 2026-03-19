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

type AppSidebarProps = React.ComponentProps<typeof Sidebar>;

export function AppSidebar({ ...props }: AppSidebarProps) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <ProjectSwitcher />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMain} />
      </SidebarContent>
      <SidebarFooter>
        <UpdateBanner />
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
