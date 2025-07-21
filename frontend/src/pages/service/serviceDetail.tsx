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
          <CardContent className="flex flex-col gap-2">
            {incidentsData.length === 0 && (
              <div className="flex items-center justify-center h-full">
                <p className="text-muted-foreground text-sm font-medium">
                  No incidents found
                </p>
              </div>
            )}
            {incidentsData?.map((incident: Incident) => (
              <div key={incident.id}>
                <div className="flex items-center justify-between flex-col md:flex-row">
                  <span className="text-muted-foreground">#{incident.id}</span>
                  <div className="flex items-center gap-2">
                    <span
                      className={cn(
                        "text-muted-foreground text-sm font-medium",
                        incident.resolved && "text-emerald-500",
                        !incident.resolved && "text-rose-500"
                      )}
                    >
                      {incident.resolved ? "Resolved" : "Active"}
                    </span>
                    <Button
                      size="sm"
                      variant="destructive"
                      className="p-0"
                      onClick={() => setDeleteIncident(incident)}
                    >
                      <TrashIcon className="size-3" />
                    </Button>
                  </div>
                </div>
                <Alert variant="destructive" className="mt-3">
                  <CircleAlertIcon />
                  <AlertDescription>
                    <div dangerouslySetInnerHTML={{ __html: incident.error }} />
                  </AlertDescription>
                </Alert>
                <div className="flex flex-col gap-2 incident-details">
                  <div className="flex items-center gap-2">
                    <strong>Start:</strong>{" "}
                    {new Date(incident.start_time).toLocaleString()}
                  </div>
                  <div className="flex items-center gap-2">
                    <strong>End:</strong>{" "}
                    {incident.end_time
                      ? new Date(incident.end_time).toLocaleString()
                      : "Still Active"}
                  </div>

                  {incident.duration ? (
                    <div className="flex items-center gap-2">
                      <strong>Duration:</strong>{" "}
                      {Number(incident.duration / 1000000000).toFixed(2)}s
                    </div>
                  ) : (
                    ""
                  )}
                </div>
                <hr className="my-4" />
              </div>
            ))}
            {incidentsCount != null && incidentsCount > filters.pageSize && (
              <PaginationTable
                className="px-0"
                selectedRows={filters.pageSize}
                setSelectedRows={(value) => setFilters({ pageSize: value })}
                selectedPage={filters.page}
                setSelectedPage={(value) => setFilters({ page: value })}
                totalPages={Math.ceil((incidentsCount ?? 0) / filters.pageSize)}
              />
            )}
          </CardContent>
        </Card>
      </div>
    </ContentWrapper>
  );
};

export default ServiceDetail;
