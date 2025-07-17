import {cn} from "@/shared/lib/utils";
import {Loader2} from "lucide-react";

interface LoaderProps {
  className?: string;
  loaderPage?: boolean;
  size?: number;
}

export const Loader = ({
  className,
  loaderPage = false,
  size = 10,
}: LoaderProps) => {
  return (
    <div
      className={cn(
        "flex justify-center items-center",
        loaderPage && "h-screen"
      )}
    >
      <div className={cn("animate-spin rounded-full", className)}>
        <Loader2 className={cn("text-muted-foreground", `size-${size}`)} />
      </div>
    </div>
  );
};
