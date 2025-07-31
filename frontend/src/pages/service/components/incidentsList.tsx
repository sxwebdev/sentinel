import {useState} from "react";
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
import {CheckIcon, CircleAlertIcon, TrashIcon, CopyIcon} from "lucide-react";
import {cn} from "@/shared/lib/utils";
import {ExpandableText} from "@/shared/components/expandableText";
import {formatDuration} from "@/shared/utils";
import PaginationTable from "@/shared/components/paginationTable";
import type {
  DbutilsFindResponseWithCountWebIncident,
  GetServicesIdIncidentsParams,
  WebIncident,
} from "@/shared/types/model";

interface IncidentsListProps {
  incidentsData: DbutilsFindResponseWithCountWebIncident;
  incidentsCount: number | null;
  filters: GetServicesIdIncidentsParams;
  setFilters: (filters: Partial<GetServicesIdIncidentsParams>) => void;
  setDeleteIncident: (incident: WebIncident) => void;
}

export const IncidentsList = ({
  incidentsData,
  incidentsCount,
  filters,
  setFilters,
  setDeleteIncident,
}: IncidentsListProps) => {
  // State to track copied incident IDs
  const [copiedIncidents, setCopiedIncidents] = useState<Set<string>>(
    new Set()
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
      console.error("Failed to copy incident ID: ", err);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Incidents</CardTitle>
      </CardHeader>
      <CardContent>
        {incidentsData.items?.length === 0 ? (
          <div className="text-center py-12">
            <CircleAlertIcon className="mx-auto h-12 w-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No incidents found</p>
          </div>
        ) : (
          <div className="space-y-3">
            {incidentsData?.items?.map((incident: WebIncident) => (
              <div
                key={incident.id}
                className="flex items-center gap-4 p-4 rounded-lg border bg-card hover:shadow-sm transition-shadow"
              >
                {/* Status Indicator */}
                <div className="flex-shrink-0">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <div
                          className={cn(
                            "w-2.5 h-2.5 rounded-full",
                            incident.resolved ? "bg-emerald-400" : "bg-rose-400"
                          )}
                        />
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>
                          {incident.resolved
                            ? "Incident resolved"
                            : "Active incident"}
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>

                {/* Main Content */}
                <div className="flex-1 min-w-0 space-y-2">
                  <div className="flex items-center gap-2 text-sm flex-wrap">
                    <TooltipProvider delayDuration={0}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <button
                            onClick={() =>
                              handleCopyIncidentId(incident.id ?? "")
                            }
                            className="font-medium text-foreground hover:text-blue-600 transition-colors duration-200 inline-flex items-center gap-1.5 px-1 py-0.5 rounded hover:bg-blue-50 cursor-pointer"
                            aria-label={
                              copiedIncidents.has(incident.id ?? "")
                                ? "Copied"
                                : "Copy incident ID"
                            }
                          >
                            #{incident.id?.slice(-6)}
                            <div className="w-3.5 h-3.5 flex items-center justify-center">
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
                        <TooltipContent className="px-2 py-1 text-xs">
                          {copiedIncidents.has(incident.id ?? "")
                            ? "Copied!"
                            : "Click to copy ID"}
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>

                    <Badge
                      variant={incident.resolved ? "default" : "destructive"}
                      className={cn(
                        "text-xs font-medium",
                        incident.resolved && "bg-emerald-100 text-emerald-600",
                        !incident.resolved && "bg-rose-100 text-rose-600"
                      )}
                    >
                      {incident.resolved ? "Resolved" : "Active"}
                    </Badge>
                  </div>

                  <div className="text-sm text-muted-foreground">
                    <ExpandableText
                      content={incident?.message ?? ""}
                      className="text-sm text-muted-foreground"
                    />
                  </div>

                  <div className="flex flex-wrap gap-2 md:gap-4 text-xs text-muted-foreground">
                    <div>
                      <span className="font-medium">Started:</span>{" "}
                      {new Date(incident?.started_at ?? "").toLocaleDateString()}{" "}
                      at{" "}
                      {new Date(incident?.started_at ?? "").toLocaleTimeString()}
                    </div>
                    {incident?.resolved_at && (
                      <div>
                        <span className="font-medium">Ended:</span>{" "}
                        {new Date(incident?.resolved_at ?? "").toLocaleDateString()} at{" "}
                        {new Date(incident?.resolved_at ?? "").toLocaleTimeString()}
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
                    onClick={() => setDeleteIncident(incident)}
                  >
                    <TrashIcon className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            ))}

            {incidentsCount != null && (
              <div className="pt-4">
                <PaginationTable
                  className="px-0"
                  selectedRows={filters.page_size ?? 0}
                  setSelectedRows={(value) => setFilters({page_size: value})}
                  selectedPage={filters.page ?? 0}
                  setSelectedPage={(value) => setFilters({page: value})}
                  totalPages={Math.ceil(
                    (incidentsCount ?? 0) / (filters.page_size ?? 0)
                  )}
                />
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};
