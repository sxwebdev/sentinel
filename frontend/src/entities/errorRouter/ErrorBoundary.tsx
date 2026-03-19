import { Component, type ReactNode } from "react";
import { Button } from "@/shared/components/ui";
import { ConnectError } from "@connectrpc/connect";

export type ErrorData = {
  status: number;
  name?: string;
  message: string;
  details?: string;
};

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
}

export class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = {
    error: null,
  };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  handleReload = () => {
    this.setState({ error: null });
    window.location.reload();
  };

  render() {
    if (this.state.error) {
      const errorData: ErrorData = {
        status: 500,
        name: this.state.error.name,
        message: this.state.error.message,
        // details: this.state.error.details,
      };

      if (this.state.error instanceof ConnectError) {
        errorData.status = 500;
        errorData.name = this.state.error.name;
        errorData.message =
          this.state.error.rawMessage ||
          this.state.error.message ||
          "Unknown error";
        errorData.details = this.state.error.details.join(", ") || undefined;
      }

      if (errorData.message == "Failed to fetch") {
        errorData.status = 503;
        errorData.name = "Service Unavailable";
        errorData.message =
          "The server is currently unreachable. Please try again later.";
        errorData.details = undefined;
      }

      return (
        <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
          <div className="w-full max-w-sm">
            <div className="flex flex-col items-center gap-4 py-10 text-center">
              <div className="text-4xl font-bold">{errorData.status}</div>
              <div className="text-2xl font-semibold">
                {errorData.name || "Something went wrong"}
              </div>
              <div className="text-muted-foreground max-w-md text-sm break-words">
                {errorData.message}
                {errorData.details && (
                  <div className="mt-2 text-xs opacity-75">
                    {errorData.details}
                  </div>
                )}
              </div>
              <div className="flex gap-2">
                <Button onClick={this.handleReload}>Reload page</Button>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
