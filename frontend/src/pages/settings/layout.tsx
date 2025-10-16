import {
  Link,
  Outlet,
  linkOptions,
  useLocation,
  useNavigate,
} from "@tanstack/react-router";
import { BellIcon, OctagonAlertIcon, SettingsIcon } from "lucide-react";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/shared/components/ui/card";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/components/ui/select";
import { Separator } from "@/shared/components/ui/separator";
import { cn } from "@/shared/lib/utils";
import { buttonVariants } from "@/shared/components/ui/button";

const menuItems = [
  linkOptions({
    to: "/settings",
    label: (
      <>
        <SettingsIcon size={16} /> General
      </>
    ),
    activeOptions: { exact: true },
  }),
  linkOptions({
    to: "/settings/notifications",
    label: (
      <>
        <BellIcon size={16} /> Notifications
      </>
    ),
  }),
  linkOptions({
    to: "/settings/notifications-history",
    label: (
      <>
        <OctagonAlertIcon size={16} /> Notifications history
      </>
    ),
  }),
];

const SettingsLayout = () => {
  const navigate = useNavigate();
  const { pathname } = useLocation();

  // Determine the closest matching menu item for the current pathname
  const selectedTo =
    menuItems
      .map((item) => item.to)
      .sort((a, b) => b.length - a.length)
      .find((to) => pathname === to || pathname.startsWith(to + "/")) ??
    "/settings";

  return (
    <Card className="gap-3">
      <CardHeader>
        <CardTitle className="text-2xl">Settings</CardTitle>
        <CardDescription>Manage application settings</CardDescription>
        <Separator className="my-2 hidden md:block" />
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-3.5 md:flex-row md:gap-5 lg:gap-12">
          <aside className="min-w-40 md:max-w-52">
            {/* Mobile View */}
            <div className="md:hidden">
              <Select
                onValueChange={(value) => navigate({ to: value })}
                value={selectedTo}
              >
                <SelectTrigger className="my-3.5 w-full">
                  <SelectValue placeholder="Select page" />
                </SelectTrigger>
                <SelectContent>
                  {menuItems.map((item) => (
                    <SelectItem key={item.to} value={item.to}>
                      <span className="flex items-center gap-2 truncate">
                        {item.label}
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Separator />
            </div>

            {/* Desktop View */}
            <nav className="sticky top-22 hidden gap-1 md:grid">
              {menuItems.map((option) => (
                <Link
                  key={option.to}
                  {...option}
                  className={cn(
                    buttonVariants({ variant: "ghost" }),
                    "flex items-center justify-start gap-3 truncate duration-50",
                    "hover:bg-accent/50",
                  )}
                  activeProps={{ className: `bg-muted hover:bg-accent/70` }}
                >
                  {option.label}
                </Link>
              ))}
            </nav>
          </aside>
          <div className="min-w-0 flex-1">
            <Outlet />
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default SettingsLayout;
