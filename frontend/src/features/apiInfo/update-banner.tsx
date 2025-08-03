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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/shared/components/ui/alert-dialog";
import Markdown from "react-markdown";
import { BookOpenText, LoaderCircleIcon } from "lucide-react";
import remarkGfm from "remark-gfm";
import { useServerStore } from "@/pages/dashboard/store/useServerStore";

export const UpdateBanner = () => {
  const serverStore = useServerStore();

  if (!serverStore.isUpdateAvailable) return null;

  return (
    <div className="bg-muted px-4 py-3 md:py-2 rounded-md mb-4">
      <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
        <p className="text-sm">
          <span className="font-medium">
            {serverStore.serverInfo?.available_update?.tag_name}
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
                  {serverStore.serverInfo?.available_update?.tag_name}
                </span>
              </h2>
              <Markdown remarkPlugins={[remarkGfm]}>
                {serverStore.serverInfo?.available_update?.description}
              </Markdown>
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline">Close</Button>
              </DialogClose>

              <Button variant="outline">
                <a
                  href={serverStore.serverInfo?.available_update?.url}
                  target="_blank"
                >
                  Release details
                </a>
              </Button>

              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="default" disabled={serverStore.isUpdating}>
                    {serverStore.isUpdating && (
                      <LoaderCircleIcon
                        className="-ms-1 animate-spin"
                        size={16}
                        aria-hidden="true"
                      />
                    )}
                    Upgrade
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This will upgrade the server to the latest version. The
                      server will restart, and you will be redirected to the
                      dashboard after the upgrade is complete.
                      <br />
                      <br />
                      Please ensure you have a backup of your data before
                      proceeding.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={() => serverStore.doUpgrade()}>
                      Let's do it!
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};
