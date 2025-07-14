import {cn} from "@/shared/lib/utils";
import {Loader2} from "lucide-react";

interface LoaderProps {
  className?: string;
  loaderPage?: boolean;
}

export const Loader = ({className, loaderPage = false}: LoaderProps) => {
  return (
    <div
      className={cn(
        "flex justify-center items-center",
        loaderPage && "h-screen"
      )}
    >
      <div className={cn("animate-spin rounded-full", className)}>
        <Loader2 className={cn("text-muted-foreground size-10")} />
      </div>
    </div>
  );
};
