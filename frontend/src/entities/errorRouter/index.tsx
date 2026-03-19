import { useRouter, type ErrorRouteComponent } from "@tanstack/react-router";
import { Button } from "@/shared/components/ui";
import { ConnectError } from "@connectrpc/connect";

export type ErrorData = {
  status: number;
  message: string;
  details?: string;
};

// TanStack Router provides the error as a generic Error. We narrow to AxiosError when possible.
const ErrorRouter: ErrorRouteComponent = ({ error, reset }) => {
  const router = useRouter();

  const errorData: ErrorData = {
    status: 500,
    message: error.name,
    details: error.message,
  };

  if (error instanceof ConnectError) {
    errorData.status = error.code;
    errorData.message = error.rawMessage || error.message || "Unknown error";
    errorData.details = error.details.join(", ") || undefined;
  }

  const handleReload = async () => {
    // Prefer built-in reset if provided (clears error boundary), then invalidate route data
    if (reset) {
      reset();
    }
    await router.invalidate();
  };

  return (
    <div className="flex flex-col items-center gap-4 py-10 text-center">
      <div className="text-4xl font-bold">{errorData.status}</div>
      <div className="text-2xl font-semibold">Something went wrong</div>
      <div className="text-muted-foreground max-w-md text-sm break-words">
        {errorData.message}
        {errorData.details && (
          <div className="mt-2 text-xs opacity-75">{errorData.details}</div>
        )}
      </div>
      <div className="flex gap-2">
        <Button onClick={handleReload}>Reload page</Button>
      </div>
    </div>
  );
};

export default ErrorRouter;
