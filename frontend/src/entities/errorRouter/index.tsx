import { useRouter, type ErrorRouteComponent } from "@tanstack/react-router";
import { Button } from "@/shared/components/ui";
import { isAxiosError, type AxiosError } from "axios";
import type { WebErrorResponse } from "@/shared/types/model/webErrorResponse";

type errorData = {
  status: number;
  message: string;
  details?: string;
};

// TanStack Router provides the error as a generic Error. We narrow to AxiosError when possible.
const ErrorRouter: ErrorRouteComponent = ({ error, reset }) => {
  const router = useRouter();

  const errorData: errorData = {
    status: 500,
    message: error.name,
    details: error.message,
  };

  const axiosErr: AxiosError<WebErrorResponse> | undefined = isAxiosError(error)
    ? error
    : undefined;

  if (axiosErr) {
    errorData.status = axiosErr.response?.status || 500;
    errorData.message = axiosErr.message || "Unknown error";
    errorData.details = axiosErr.response?.data.error;
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
