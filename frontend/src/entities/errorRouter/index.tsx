import { useRouter, type ErrorRouteComponent } from "@tanstack/react-router";
import { Button } from "@/shared/components/ui";
import { isAxiosError, type AxiosError } from "axios";

// TanStack Router provides the error as a generic Error. We narrow to AxiosError when possible.
const ErrorRouter: ErrorRouteComponent = ({ error, reset }) => {
  const router = useRouter();

  const axiosErr: AxiosError | undefined = isAxiosError(error)
    ? error
    : undefined;

  // Try to extract a numeric status; fall back to 500
  const status =
    typeof axiosErr?.response?.status === "number"
      ? axiosErr.response!.status
      : typeof (error as { status?: unknown })?.status === "number"
        ? (error as { status?: number }).status!
        : 500;

  const message = axiosErr?.message || error.message || "Unknown error";

  const details: unknown = axiosErr?.response?.data;

  const handleReload = async () => {
    // Prefer built-in reset if provided (clears error boundary), then invalidate route data
    if (reset) {
      reset();
    }
    await router.invalidate();
  };

  return (
    <div className="flex flex-col items-center gap-4 py-10 text-center">
      <div className="text-4xl font-bold">{status}</div>
      <div className="text-2xl font-semibold">Something went wrong</div>
      <div className="text-muted-foreground max-w-md text-sm break-words">
        {message}
        {details && typeof details === "object" && "error" in details ? (
          <div className="mt-2 text-xs opacity-75">
            {(() => {
              const errVal = (details as { error: unknown }).error;
              return typeof errVal === "string"
                ? errVal
                : JSON.stringify(errVal);
            })()}
          </div>
        ) : null}
      </div>
      <div className="flex gap-2">
        <Button onClick={handleReload}>Reload page</Button>
      </div>
    </div>
  );
};

export default ErrorRouter;
