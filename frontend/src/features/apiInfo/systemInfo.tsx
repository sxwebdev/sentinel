import { getSystemInfo } from "@/api/gen/sentinel/system/v1/service-SystemService_connectquery";
import { useQuery } from "@connectrpc/connect-query";
import { Button } from "@shared/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@shared/components/ui/popover";
import { Info } from "lucide-react";

export const SystemInfo = () => {
  const q = useQuery(getSystemInfo);

  if (!q.data) return null;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button size="icon" variant="ghost" aria-label="Server info">
          <Info size={16} className="size-5" aria-hidden="true" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="max-w-[180px] py-3 shadow-none" side="top">
        <ul className="grid gap-3 text-sm">
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Sentinel version</span>
            <span className="flex items-center justify-between font-medium">
              {q.data.info?.version}
            </span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Go version</span>
            <span className="font-medium">{q.data.info?.goVersion}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">SQLite version</span>
            <span className="font-medium">{q.data.info?.sqliteVersion}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">OS</span>
            <span className="font-medium">{q.data.info?.os}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Arch</span>
            <span className="font-medium">{q.data.info?.arch}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Build date</span>
            <span className="font-medium">{q.data.info?.buildDate}</span>
          </li>
          <li className="grid gap-0.5">
            <span className="text-muted-foreground">Commit hash</span>
            <span className="font-medium">
              {q.data.info?.commitHash?.slice(0, 8) || "N/A"}
            </span>
          </li>
        </ul>
      </PopoverContent>
    </Popover>
  );
};
