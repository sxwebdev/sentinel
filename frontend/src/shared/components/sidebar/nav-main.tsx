"use client";

import { Link, type LinkOptions } from "@tanstack/react-router";

import { type LucideIcon } from "lucide-react";

import { Collapsible } from "@/shared/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/shared/components/ui/sidebar";
import { cn } from "@/shared/lib/utils";

type NavMainItem = LinkOptions & {
  to: string;
  title: string;
  icon?: LucideIcon;
};

type NavMainProps = {
  items: {
    main: NavMainItem & { isActive?: boolean };
    items?: NavMainItem[];
  }[];
};

export function NavMain({ items }: NavMainProps) {
  return (
    <SidebarGroup>
      {/* <SidebarGroupLabel>Application</SidebarGroupLabel> */}
      <SidebarMenu>
        {items.map((option) => (
          <Collapsible
            key={option.main.title}
            asChild
            defaultOpen={option.main.isActive}
          >
            <SidebarMenuItem>
              <SidebarMenuButton asChild tooltip={option.main.title}>
                <Link
                  key={option.main.to}
                  {...option}
                  to={option.main.to}
                  className="text-sm/6 font-medium"
                  activeProps={{
                    className: cn(
                      "bg-sidebar-accent text-sidebar-accent-foreground hover:bg-sidebar-accent/70",
                    ),
                  }}
                >
                  {option.main.icon && <option.main.icon />}
                  <span>{option.main.title}</span>
                </Link>
              </SidebarMenuButton>
              {/* {option.items?.length ? (
                <>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuAction className="data-[state=open]:rotate-90">
                      <ChevronRight />
                      <span className="sr-only">Toggle</span>
                    </SidebarMenuAction>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {option.items?.map((subOption) => (
                        <SidebarMenuSubItem key={subOption.to}>
                          <SidebarMenuSubButton asChild>
                            <Link
                              key={subOption.to}
                              {...subOption}
                              to={subOption.to}
                              className="text-sm/6"
                              activeProps={{
                                className: cn(
                                  "bg-sidebar-accent text-sidebar-accent-foreground hover:bg-sidebar-accent/70",
                                ),
                              }}
                            >
                              {subOption.icon && <subOption.icon />}
                              <span>{subOption.title}</span>
                            </Link>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      ))}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </>
              ) : null} */}
            </SidebarMenuItem>
          </Collapsible>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
