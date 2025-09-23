import { useState } from "react";
import { getIncidents } from "@/shared/api/gen/incidents/incidents";
import { ExpandableText } from "@/shared/components/expandableText";
import PaginationTable from "@/shared/components/paginationTable";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/shared/components/ui";
import { cn } from "@/shared/lib/utils";
import type {
  StorecmnFindResponseWithCountModelsIncident,
  GetIncidentsParams,
  ModelsIncident,
  WebErrorResponse,
} from "@/shared/types/model";
import { formatDuration } from "@/shared/utils/duration";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { CheckIcon, CircleAlertIcon, CopyIcon, TrashIcon } from "lucide-react";
import { toast } from "sonner";
import type { AxiosError } from "axios";

export const Route = createFileRoute("/incidents")({
  component: RouteComponent,
  loaderDeps: ({
    search: { page = 1, page_size = 10 },
  }: {
    search: GetIncidentsParams;
  }) => ({
    page,
    page_size,
  }),
  loader: ({ deps: { page, page_size } }) =>
    getIncidents().getIncidents({ page, page_size }),
  gcTime: 0,
});

function RouteComponent() {
  const data = Route.useLoaderData();

  return <IncidentsList incidentsData={data} />;
}

interface IncidentsListProps {
  incidentsData: StorecmnFindResponseWithCountModelsIncident;
}

export const IncidentsList = ({ incidentsData }: IncidentsListProps) => {
  const deps = Route.useLoaderDeps();
  const nav = Route.useNavigate();
  const router = useRouter();

  // State to track copied incident IDs
  const [copiedIncidents, setCopiedIncidents] = useState<Set<string>>(
    new Set(),
  );

  const handleCopyIncidentId = async (incidentId: string) => {
    try {
      await navigator.clipboard.writeText(incidentId);
      setCopiedIncidents((prev) => new Set(prev).add(incidentId));
      setTimeout(() => {
        setCopiedIncidents((prev) => {
          const newSet = new Set(prev);
          newSet.delete(incidentId);
          return newSet;
        });
      }, 1500);
    } catch (err) {
      toast.error("Failed to copy incident ID", {
        description: (err as Error).message,
      });
    }
  };

  const handleDeleteIncident = async (
    serviceId: string,
    incidentId: string,
  ) => {
    try {
      await getIncidents().deleteServicesIdIncidentsIncidentId(
        serviceId,
        incidentId,
      );
      router.invalidate();
      toast.success("Incident deleted", { description: `ID: ${incidentId}` });
    } catch (err: unknown) {
      toast.error("Failed to delete incident", {
        description:
          (err as AxiosError<WebErrorResponse>)?.response?.data?.error ||
          (err as Error).message,
      });
    }
  };

  return (
    <Card className="gap-3">
      <CardHeader>
        <CardTitle className="pb-0 text-xl">Recent Incidents</CardTitle>
      </CardHeader>
      <CardContent>
        {incidentsData.items?.length === 0 ? (
          <div className="py-12 text-center">
            <CircleAlertIcon className="text-muted-foreground/50 mx-auto mb-4 h-12 w-12" />
            <p className="text-muted-foreground">No incidents found</p>
          </div>
        ) : (
          <div className="space-y-3">
            {incidentsData?.items?.map((incident: ModelsIncident) => (
              <div
                key={incident.id}
                className="bg-card flex items-center gap-4 rounded-lg border p-4 transition-shadow hover:shadow-sm"
              >
                {/* Status Indicator */}
                <div className="flex-shrink-0">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div
                          className={cn(
                            "h-2.5 w-2.5 rounded-full",
                            incident.resolved_at
                              ? "bg-emerald-400"
                              : "bg-rose-400",
                          )}
                        />
                      </TooltipTrigger>
                      <TooltipContent showArrow className="dark">
                        <p>
                          {incident.resolved_at
                            ? "Incident resolved"
                            : "Active incident"}
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>

                {/* Main Content */}
                <div className="min-w-0 flex-1 space-y-2">
                  <div className="flex flex-wrap items-center gap-2 text-sm">
                    <TooltipProvider delayDuration={0}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <button
                            onClick={() =>
                              handleCopyIncidentId(incident.id ?? "")
                            }
                            className="text-foreground inline-flex cursor-pointer items-center gap-1.5 rounded px-1 py-0.5 font-medium transition-colors duration-200 hover:bg-blue-50 hover:text-blue-600"
                            aria-label={
                              copiedIncidents.has(incident.id ?? "")
                                ? "Copied"
                                : "Copy incident ID"
                            }
                          >
                            #{incident.id?.slice(-6)}
                            <div className="flex h-3.5 w-3.5 items-center justify-center">
                              {copiedIncidents.has(incident.id ?? "") ? (
                                <CheckIcon
                                  className="stroke-emerald-500 transition-all duration-200"
                                  aria-hidden="true"
                                />
                              ) : (
                                <CopyIcon
                                  aria-hidden="true"
                                  className="transition-all duration-200"
                                />
                              )}
                            </div>
                          </button>
                        </TooltipTrigger>
                        <TooltipContent
                          showArrow
                          className="dark px-2 py-1 text-xs"
                        >
                          {copiedIncidents.has(incident.id ?? "")
                            ? "Copied!"
                            : "Click to copy ID"}
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>

                    <Badge
                      variant={incident.resolved_at ? "default" : "destructive"}
                      className={cn(
                        "text-xs font-medium",
                        incident.resolved_at &&
                          "bg-emerald-100 text-emerald-600",
                        !incident.resolved_at && "bg-rose-100 text-rose-600",
                      )}
                    >
                      {incident.resolved_at ? "Resolved" : "Active"}
                    </Badge>
                  </div>

                  <div className="text-muted-foreground text-sm">
                    <ExpandableText
                      content={incident?.error ?? ""}
                      className="text-muted-foreground text-sm"
                    />
                  </div>

                  <div className="text-muted-foreground flex flex-wrap gap-2 text-xs md:gap-4">
                    <div>
                      <span className="font-medium">Started:</span>{" "}
                      {new Date(
                        incident?.started_at ?? "",
                      ).toLocaleDateString()}{" "}
                      at{" "}
                      {new Date(
                        incident?.started_at ?? "",
                      ).toLocaleTimeString()}
                    </div>
                    {incident?.resolved_at && (
                      <div>
                        <span className="font-medium">Ended:</span>{" "}
                        {new Date(
                          incident?.resolved_at ?? "",
                        ).toLocaleDateString()}{" "}
                        at{" "}
                        {new Date(
                          incident?.resolved_at ?? "",
                        ).toLocaleTimeString()}
                      </div>
                    )}
                    {incident.duration && (
                      <div>
                        <span className="font-medium">Duration:</span>{" "}
                        {formatDuration(Number(incident?.duration ?? 0))}
                      </div>
                    )}
                  </div>
                </div>

                {/* Action Button */}
                <div className="flex-shrink-0">
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-8 w-8 p-0 opacity-50 hover:opacity-100"
                    onClick={() =>
                      handleDeleteIncident(
                        incident.service_id as string,
                        incident.id as string,
                      )
                    }
                  >
                    <TrashIcon className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            ))}

            <div className="pt-4">
              <PaginationTable
                className="px-0"
                selectedRows={deps.page_size ?? 0}
                setSelectedRows={(value) => {
                  nav({ search: { page_size: value, page: deps.page } });
                }}
                selectedPage={deps.page ?? 0}
                setSelectedPage={(value) => {
                  nav({ search: { page_size: deps.page_size, page: value } });
                }}
                totalPages={Math.ceil(
                  (incidentsData?.count ?? 0) / (deps.page_size ?? 0),
                )}
              />
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};
