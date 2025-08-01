import type { WebServerInfoResponse } from "@/shared/types/model";
import { Button } from "@shared/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@shared/components/ui/tooltip";
import { Info } from "lucide-react";

export const ServerInfo = ({
  serverInfo: apiInfo,
}: {
  serverInfo: WebServerInfoResponse | null;
}) => {
  if (!apiInfo) return null;

  return (
    <div className="flex justify-center mt-6">
      <TooltipProvider delayDuration={0}>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="outline" size="sm">
              <Info className="text-muted-foreground" />
              <span>Server info</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent className="py-3">
            <ul className="grid gap-3 text-xs">
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">Sentinel version</span>
                <span className="font-medium">{apiInfo.version}</span>
              </li>
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">Go version</span>
                <span className="font-medium">{apiInfo.go_version}</span>
              </li>
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">SQLite version</span>
                <span className="font-medium">{apiInfo.sqlite_version}</span>
              </li>
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">OS</span>
                <span className="font-medium">{apiInfo.os}</span>
              </li>
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">Arch</span>
                <span className="font-medium">{apiInfo.arch}</span>
              </li>
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">Build date</span>
                <span className="font-medium">{apiInfo.build_date}</span>
              </li>
              <li className="grid gap-0.5">
                <span className="text-muted-foreground">Commit hash</span>
                <span className="font-medium">{apiInfo.commit_hash}</span>
              </li>
            </ul>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
};
