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
import { LoaderCircleIcon, SparklesIcon } from "lucide-react";
import remarkGfm from "remark-gfm";
import { useServerStore } from "@/pages/dashboard/store/useServerStore";

export const UpdateBanner = () => {
  const serverStore = useServerStore();

  if (!serverStore.isUpdateAvailable) return null;

  return (
    <div className="hidden md:block">
      <Dialog>
        <DialogTrigger asChild>
          <Button size="sm" variant="ghost" className="min-w-24">
            <SparklesIcon size={16} />
            Available new update
          </Button>
        </DialogTrigger>
        <DialogContent className="flex flex-col gap-0 p-0 max-h-[85vh] sm:max-h-[min(840px,95vh)] sm:max-w-2xl [&>button:last-child]:top-3.5">
          <DialogHeader className="contents space-y-0 text-left">
            <DialogTitle className="border-b px-6 py-4 text-base">
              🚀 Available new update
            </DialogTitle>
          </DialogHeader>

          <DialogDescription asChild>
            <div className="flex-1 changelog overflow-y-auto overscroll-contain p-6">
              {/* Current version */}
              <div className="mb-3 text-lg font-semibold">
                Current version:{" "}
                <span className="text-zinc-500">
                  {serverStore.serverInfo?.version}
                </span>
              </div>

              {/* New version */}
              <Markdown remarkPlugins={[remarkGfm]}>
                {serverStore.serverInfo?.available_update?.description}
              </Markdown>
            </div>
          </DialogDescription>

          <DialogFooter className="flex-shrink-0 border-t px-6 py-4 sm:items-center">
            <DialogClose asChild>
              <Button variant="outline">Close</Button>
            </DialogClose>

            <Button variant="default">
              <a
                href={serverStore.serverInfo?.available_update?.url}
                target="_blank"
              >
                Release details
              </a>
            </Button>

            {serverStore.serverInfo?.available_update?.is_available_manual && (
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
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};
