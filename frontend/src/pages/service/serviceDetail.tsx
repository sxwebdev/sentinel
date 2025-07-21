import ContentWrapper from "@/widgets/wrappers/contentWrapper";
import { useServiceDetail } from "./hooks/useServiceDetail";
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
import {
  ArrowLeftIcon,
  CheckIcon,
  CircleAlertIcon,
  PlayIcon,
  TrashIcon,
  CopyIcon,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import { useIsMobile } from "@/shared/hooks/useIsMobile";
import { cn } from "@/shared/lib/utils";
import { InfoCardStats } from "@/entities/infoStatsCard/infoCardStats";
import type { Incident } from "../../features/service/types/type";
import { Loader } from "@/entities/loader/loader";
import { ConfirmDialog } from "@/entities/confirmDialog/confirmDialog";
import { ActivityIndicatorSVG } from "@/entities/ActivityIndicatorSVG/ActivityIndicatorSVG";
import { Link } from "react-router";
import PaginationTable from "@/shared/components/paginationTable";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/shared/components/ui/alert";
import { useState } from "react";

const ServiceDetail = () => {
  const {
    filters,
    incidentsData,
    incidentsCount,
    deleteIncident,
    resolveIncident,
    serviceDetailData,
    serviceStatsData,
    setFilters,
    onCheckService,
    setDeleteIncident,
    onDeleteIncident,
    setResolveIncident,
    onResolveIncident,
  } = useServiceDetail();
  const isMobile = useIsMobile();

  // State to track expanded incident errors
  const [expandedIncidents, setExpandedIncidents] = useState<Set<string>>(
    new Set()
  );

  // State to track copied incident IDs
  const [copiedIncidents, setCopiedIncidents] = useState<Set<string>>(
    new Set()
  );

  const toggleIncidentExpansion = (incidentId: string) => {
    setExpandedIncidents((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(incidentId)) {
        newSet.delete(incidentId);
      } else {
        newSet.add(incidentId);
      }
      return newSet;
    });
  };

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

  // Helper function to format duration from nanoseconds to human readable format
  const formatDuration = (nanoseconds: number) => {
    const seconds = Math.floor(nanoseconds / 1000000000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) {
      const remainingHours = hours % 24;
      const remainingMinutes = minutes % 60;
      return `${days}d ${remainingHours}h ${remainingMinutes}m`;
    } else if (hours > 0) {
      const remainingMinutes = minutes % 60;
      const remainingSeconds = seconds % 60;
      return `${hours}h ${remainingMinutes}m ${remainingSeconds}s`;
    } else if (minutes > 0) {
      const remainingSeconds = seconds % 60;
      return `${minutes}m ${remainingSeconds}s`;
    } else {
      return `${seconds}s`;
    }
  };

  if (!serviceDetailData || !incidentsData || !serviceStatsData)
    return <Loader loaderPage />;

  const cardStats = [
    {
      value: serviceDetailData?.total_incidents,
      key: "total_incidents",
      description: "Total Incidents",
    },
    {
      value: serviceDetailData?.total_checks,
      key: "total_checks",
      description: "Total Checks",
    },
    {
      value: `${(serviceStatsData?.avg_response_time / 1000000).toFixed(1)} ms`,
      key: "avg_response_time",
      description: "Avg Response Time",
    },
    {
      value: `${serviceStatsData?.uptime_percentage.toFixed(1)}%`,
      key: "uptime",
      description: "Uptime",
    },
    {
      value: serviceDetailData?.consecutive_success,
      key: "consecutive_success",
      description: "Consecutive Success",
    },
    {
      value: serviceDetailData?.consecutive_fails,
      key: "consecutive_fails",
      description: "Consecutive Fails",
    },
  ];

  return (
    <ContentWrapper>
      <ConfirmDialog
        open={resolveIncident}
        setOpen={() => setResolveIncident(false)}
        onSubmit={onResolveIncident}
        title="Resolve Incident"
        description="Are you sure you want to resolve this incident?"
        type="default"
      />
      <ConfirmDialog
        open={!!deleteIncident}
        setOpen={() => setDeleteIncident(null)}
        onSubmit={() => onDeleteIncident(deleteIncident?.id ?? "")}
        title="Delete Incident"
        description="Are you sure you want to delete this incident?"
        type="delete"
      />
      <div className="flex flex-col gap-6">
        <header
          className={cn(
            "flex items-center gap-2 justify-between py-2",
            isMobile && "flex-col gap-2"
          )}
        >
          <Link to={"/"}>
            <Button className="group" variant="ghost">
              <ArrowLeftIcon
                className="-ms-1 opacity-60 transition-transform group-hover:-translate-x-0.5"
                size={16}
                aria-hidden="true"
              />
              Back
            </Button>
          </Link>
          <div
            className={cn(
              "flex items-center gap-2",
              isMobile && "w-full flex-col"
            )}
          >
            <Button
              size="sm"
              className={cn(isMobile && "w-full")}
              onClick={() => onCheckService(serviceDetailData?.id)}
            >
              <PlayIcon />
              Trigger Check
            </Button>
            <Button
              size="sm"
              variant="outline"
              className={cn(isMobile && "w-full")}
              onClick={() => setResolveIncident(true)}
            >
              <CheckIcon />
              Resolve Incidents
            </Button>
          </div>
        </header>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-3.5 md:gap-4">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <ActivityIndicatorSVG
                      active={serviceDetailData?.is_enabled}
                      size={24}
                    />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>
                      {serviceDetailData?.is_enabled
                        ? "Service enabled"
                        : "Service disabled"}
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <div>
                <h3 className="text-base md:text-lg font-bold">
                  {serviceDetailData?.name}
                </h3>
                <div className="mt-2">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge variant={"secondary"} className="text-sm">
                          {serviceDetailData?.protocol}
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>Protocol</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge variant={"secondary"} className="ml-3 text-sm">
                          {new Date(
                            serviceDetailData?.last_check
                          ).toLocaleString()}
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>Last Check</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </div>
              <Badge
                className={cn(
                  "text-xs md:text-sm font-semibold ml-auto",
                  serviceDetailData?.status === "up" &&
                    "bg-emerald-100 text-emerald-600",
                  serviceDetailData?.status === "down" &&
                    "bg-rose-100 text-rose-600",
                  serviceDetailData?.status === "unknown" &&
                    "bg-yellow-100 text-yellow-600"
                )}
              >
                {serviceDetailData?.status.toLocaleUpperCase()}
              </Badge>
            </CardTitle>
          </CardHeader>
        </Card>
        {serviceDetailData.last_error && (
          <Alert variant="destructive">
            <CircleAlertIcon />
            <AlertTitle className="font-semibold">Last Error</AlertTitle>
            <AlertDescription>
              <div
                dangerouslySetInnerHTML={{
                  __html: serviceDetailData.last_error,
                }}
              />
            </AlertDescription>
          </Alert>
        )}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {cardStats.map((stat) => (
            <InfoCardStats
              key={stat.key}
              title={stat.description}
              value={stat?.value?.toString() ?? "0"}
            />
          ))}
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Recent Incidents</CardTitle>
          </CardHeader>
          <CardContent>
            {incidentsData.length === 0 ? (
              <div className="text-center py-12">
                <CircleAlertIcon className="mx-auto h-12 w-12 text-muted-foreground/50 mb-4" />
                <p className="text-muted-foreground">No incidents found</p>
              </div>
            ) : (
              <div className="space-y-3">
                {incidentsData?.map((incident: Incident) => (
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
                                incident.resolved
                                  ? "bg-emerald-400"
                                  : "bg-rose-400"
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
                                  handleCopyIncidentId(incident.id)
                                }
                                className="font-medium text-foreground hover:text-blue-600 transition-colors duration-200 inline-flex items-center gap-1.5 px-1 py-0.5 rounded hover:bg-blue-50 cursor-pointer"
                                aria-label={
                                  copiedIncidents.has(incident.id)
                                    ? "Copied"
                                    : "Copy incident ID"
                                }
                              >
                                #{incident.id.slice(-6)}
                                <div className="w-3.5 h-3.5 flex items-center justify-center">
                                  {copiedIncidents.has(incident.id) ? (
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
                              {copiedIncidents.has(incident.id)
                                ? "Copied!"
                                : "Click to copy ID"}
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>

                        <Badge
                          variant={
                            incident.resolved ? "default" : "destructive"
                          }
                          className={cn(
                            "text-xs font-medium",
                            incident.resolved &&
                              "bg-emerald-100 text-emerald-600",
                            !incident.resolved && "bg-rose-100 text-rose-600"
                          )}
                        >
                          {incident.resolved ? "Resolved" : "Active"}
                        </Badge>
                      </div>

                      <div className="text-sm text-muted-foreground">
                        <div
                          className={cn(
                            !expandedIncidents.has(incident.id) &&
                              "line-clamp-1"
                          )}
                          dangerouslySetInnerHTML={{ __html: incident.error }}
                        />
                        {/* Show toggle link if content is likely to be truncated or on mobile */}
                        {incident.error &&
                          (incident.error.length > 80 || isMobile) && (
                            <button
                              onClick={() =>
                                toggleIncidentExpansion(incident.id)
                              }
                              className="inline-flex items-center gap-1 mt-2 px-2 py-1 text-xs font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 hover:text-blue-700 rounded-md transition-colors duration-200"
                            >
                              {expandedIncidents.has(incident.id) ? (
                                <>
                                  <ChevronUp size={18} />
                                  Show less
                                </>
                              ) : (
                                <>
                                  <ChevronDown size={18} />
                                  Show more
                                </>
                              )}
                            </button>
                          )}
                      </div>

                      <div className="flex flex-wrap gap-2 md:gap-4 text-xs text-muted-foreground">
                        <div>
                          <span className="font-medium">Started:</span>{" "}
                          {new Date(incident.start_time).toLocaleDateString()}{" "}
                          at{" "}
                          {new Date(incident.start_time).toLocaleTimeString()}
                        </div>
                        {incident.end_time && (
                          <div>
                            <span className="font-medium">Ended:</span>{" "}
                            {new Date(incident.end_time).toLocaleDateString()}{" "}
                            at{" "}
                            {new Date(incident.end_time).toLocaleTimeString()}
                          </div>
                        )}
                        {incident.duration && (
                          <div>
                            <span className="font-medium">Duration:</span>{" "}
                            {formatDuration(incident.duration)}
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

                {incidentsCount != null &&
                  incidentsCount > filters.pageSize && (
                    <div className="pt-4">
                      <PaginationTable
                        className="px-0"
                        selectedRows={filters.pageSize}
                        setSelectedRows={(value) =>
                          setFilters({ pageSize: value })
                        }
                        selectedPage={filters.page}
                        setSelectedPage={(value) => setFilters({ page: value })}
                        totalPages={Math.ceil(
                          (incidentsCount ?? 0) / filters.pageSize
                        )}
                      />
                    </div>
                  )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </ContentWrapper>
  );
};

export default ServiceDetail;
