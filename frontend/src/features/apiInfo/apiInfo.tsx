import type { WebServerInfoResponse } from "@/shared/types/model";
import { Button } from "@shared/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@shared/components/ui/tooltip";

export const ApiInfo = ({
  apiInfo,
}: {
  apiInfo: WebServerInfoResponse | null;
}) => {
  if (!apiInfo) return null;

  return (
    <TooltipProvider delayDuration={0}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="outline" size="sm">
            <span>
              <span className="text-muted-foreground">Version:</span>{" "}
              {apiInfo.version}
            </span>
          </Button>
        </TooltipTrigger>
        <TooltipContent className="py-3">
          <ul className="grid gap-3 text-xs">
            <li className="grid gap-0.5">
              <span className="text-muted-foreground">Go Version</span>
              <span className="font-medium">{apiInfo.go_version}</span>
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
              <span className="text-muted-foreground">Version</span>
              <span className="font-medium">{apiInfo.version}</span>
            </li>
            <li className="grid gap-0.5">
              <span className="text-muted-foreground">Build Date</span>
              <span className="font-medium">{apiInfo.build_date}</span>
            </li>
            <li className="grid gap-0.5">
              <span className="text-muted-foreground">Commit Hash</span>
              <span className="font-medium">{apiInfo.commit_hash}</span>
            </li>
          </ul>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};
