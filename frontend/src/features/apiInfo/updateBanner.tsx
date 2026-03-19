import {
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Badge,
} from "@/shared/components/ui";
import Markdown from "react-markdown";
import { SparklesIcon } from "lucide-react";
import remarkGfm from "remark-gfm";
import { useQuery } from "@connectrpc/connect-query";
import { checkForUpdates } from "@/api/gen/sentinel/system/v1/service-SystemService_connectquery";
import {
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/shared/components/ui/sidebar";

export const UpdateBanner = () => {
  const { data } = useQuery(checkForUpdates);

  if (!data || !data.info?.isAvailable) return null;

  return (
    <Dialog>
      <DialogTrigger asChild>
        <SidebarMenuItem>
          <SidebarMenuButton className="group/update-banner flex w-full cursor-pointer items-center justify-between">
            <div className="flex shrink-0 items-center gap-2">
              <SparklesIcon className="size-4 transition-transform duration-300 group-hover/update-banner:rotate-[90deg]" />{" "}
              <span>Available update</span>
            </div>
            <Badge variant="outline" className="bg-background">
              {data.info?.details?.tagName}
            </Badge>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </DialogTrigger>
      <DialogContent className="flex max-h-[85vh] flex-col gap-0 p-0 sm:max-h-[min(840px,95vh)] sm:max-w-2xl [&>button:last-child]:top-3.5">
        <DialogHeader className="contents space-y-0 text-left">
          <DialogTitle className="border-b px-6 py-4 text-base">
            🚀 Available new update
          </DialogTitle>
        </DialogHeader>

        <DialogDescription asChild>
          <div className="changelog flex-1 overflow-y-auto overscroll-contain p-6">
            {/* Current version */}
            <div className="mb-3 text-lg font-semibold">
              Current version:{" "}
              <span className="text-zinc-500">{data.info?.currentVersion}</span>
            </div>

            {/* New version */}
            <Markdown remarkPlugins={[remarkGfm]}>
              {data.info?.details?.description}
            </Markdown>
          </div>
        </DialogDescription>

        <DialogFooter className="flex-shrink-0 border-t px-6 py-4 sm:items-center">
          <DialogClose asChild>
            <Button variant="outline">Close</Button>
          </DialogClose>

          <Button variant="default">
            <a href={data.info?.details?.url} target="_blank">
              Release details
            </a>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
