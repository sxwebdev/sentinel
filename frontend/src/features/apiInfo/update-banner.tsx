import {
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/shared/components/ui";
import type { WebServerInfoResponse } from "@/shared/types/model";
import Markdown from "react-markdown";
import { BookOpenText } from "lucide-react";
import remarkGfm from "remark-gfm";

export const UpdateBanner = ({
  serverInfo: apiInfo,
}: {
  serverInfo: WebServerInfoResponse | null;
}) => {
  if (!apiInfo?.available_update) return null;

  return (
    <div className="bg-muted px-4 py-3 md:py-2 rounded-md mb-4">
      <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
        <p className="text-sm">
          <span className="font-medium">
            {apiInfo.available_update.tag_name}
          </span>
          <span className="text-muted-foreground mx-2">•</span>
          🚀 Available new update
        </p>
        <Dialog>
          <DialogTrigger asChild>
            <Button size="sm" variant="outline" className="min-w-24">
              <BookOpenText size={16} className="-ms-0.5" aria-hidden="true" />
              Changelog
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>🚀 Available new update</DialogTitle>
            </DialogHeader>
            <div className="changelog">
              <h2>
                Version:{" "}
                <span className="text-accent-foreground font-semibold">
                  {apiInfo.available_update.tag_name}
                </span>
              </h2>
              <Markdown remarkPlugins={[remarkGfm]}>
                {apiInfo.available_update.description}
              </Markdown>
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button>Close</Button>
              </DialogClose>

              <Button variant="outline">
                <a href={apiInfo.available_update?.url} target="_blank">
                  Release details
                </a>
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};
