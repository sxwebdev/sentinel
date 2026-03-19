import { Button } from "@/shared/components/ui";
import { ConnectError } from "@connectrpc/connect";
import React from "react";

export type ErrorData = {
  status: number;
  message: string;
  details?: string;
};

interface PageErrorProps {
  error: Error | ConnectError | ErrorData;
  onReload?: () => void;
}

export const AppError: React.FC<PageErrorProps> = ({ error, onReload }) => {
  let errorData: ErrorData = {
    status: 500,
    message: "Unknown error",
    details: undefined,
  };

  if (error instanceof ConnectError) {
    errorData.status = error.code;
    errorData.message = error.rawMessage || error.message || "Unknown error";
    errorData.details = error.details?.join(", ") || undefined;
  } else if (error instanceof Error) {
    errorData.message = error.message;
    errorData.details = error.stack;
  } else if (typeof error === "object" && error !== null) {
    errorData = { ...errorData, ...error };
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 py-10 text-center">
      <div className="text-4xl font-bold">{errorData.status}</div>
      <div className="text-2xl font-semibold">Something went wrong</div>
      <div className="text-muted-foreground max-w-md text-sm break-words">
        {errorData.message}
        {errorData.details && (
          <div className="mt-2 text-xs opacity-75">{errorData.details}</div>
        )}
      </div>
      <div className="flex gap-2">
        <Button onClick={onReload || (() => window.location.reload())}>
          Reload page
        </Button>
      </div>
    </div>
  );
};

export default AppError;
