import { useServerStore } from "@/pages/dashboard/store/useServerStore";
import { Badge } from "@/shared/components/ui";
import { Button } from "@shared/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@shared/components/ui/popover";
import { Info } from "lucide-react";

export const ServerInfo = () => {
  const serverInfo = useServerStore((s) => s.serverInfo);

  if (!serverInfo) return null;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          className="size-8 rounded-full"
          size="icon"
          variant="ghost"
          aria-label="Server info"
        >
          <Info
            size={16}
            aria-hidden="true"
            className="text-muted-foreground"
          />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="max-w-[180px] py-3 shadow-none" side="top">
        <ul className="grid gap-3 text-sm">
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Sentinel version</span>
            <span className="flex justify-between items-center font-medium">
              {serverInfo.version}
              {serverInfo.available_update ? (
                <Badge className="bg-rose-500 text-white ml-2">Outdated</Badge>
              ) : (
                <Badge className="bg-emerald-500 text-white ml-2">Latest</Badge>
              )}
            </span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Go version</span>
            <span className="font-medium">{serverInfo.go_version}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">SQLite version</span>
            <span className="font-medium">{serverInfo.sqlite_version}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">OS</span>
            <span className="font-medium">{serverInfo.os}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Arch</span>
            <span className="font-medium">{serverInfo.arch}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Build date</span>
            <span className="font-medium">{serverInfo.build_date}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Commit hash</span>
            <span className="font-medium">
              {serverInfo.commit_hash?.slice(0, 8) || "N/A"}
            </span>
          </li>
        </ul>
      </PopoverContent>
    </Popover>
  );
};
