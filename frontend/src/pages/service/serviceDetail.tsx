import ContentWrapper from "@/widgets/wrappers/contentWrapper";
import {useServiceDetail} from "./hooks/useServiceDetail";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/components/ui";
import {ArrowLeftIcon, CheckIcon, PlayIcon, TrashIcon} from "lucide-react";
import {useNavigate} from "react-router";
import {useIsMobile} from "@/shared/hooks/useIsMobile";
import {cn} from "@/shared/lib/utils";
import {InfoCardStats} from "@/entities/infoStatsCard/infoCardStats";
import type {Incident} from "./types/type";
import {Loader} from "@/entities/loader/loader";
import {ConfirmDialog} from "@/entities/confirmDialog/confirmDialog";
import { ActivityIndicatorSVG } from "@/entities/ActivityIndicatorSVG/ActivityIndicatorSVG";

const ServiceDetail = () => {
  const navigate = useNavigate();
  const {
    serviceDetailData,
    incidentsData,
    serviceStatsData,
    onCheckService,
    deleteIncident,
    setDeleteIncident,
    onDeleteIncident,
    resolveIncident,
    setResolveIncident,
    onResolveIncident,
  } = useServiceDetail();
  const isMobile = useIsMobile();

  if (!serviceDetailData || !incidentsData || !serviceStatsData)
    return <Loader loaderPage />;

  const cardStats = [
    {
      value: serviceDetailData?.service.total_incidents,
      key: "total_incidents",
      description: "Total Incidents",
    },
    {
      value: `${serviceStatsData?.uptime_percentage.toFixed(1)}%`,
      key: "uptime",
      description: "Uptime",
    },
    {
      value: `${(serviceStatsData?.avg_response_time / 1000000).toFixed(1)} ms`,
      key: "avg_response_time",
      description: "Avg Response Time",
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
        <Card className={cn("p-6", isMobile && "p-4")}>
          <header
            className={cn(
              "flex items-center gap-2 justify-between",
              isMobile && "flex-col gap-2"
            )}
          >
            <Button
              size={"sm"}
              variant="link"
              className={cn(
                "cursor-pointer  text-lg",
                isMobile && "w-full text-base"
              )}
              onClick={() => navigate("/")}
            >
              <ArrowLeftIcon />
              Back
            </Button>
            <h1 className={cn("text-2xl font-bold", isMobile && "text-lg")}>
              Service: {serviceDetailData?.service.name}
            </h1>
            <div
              className={cn(
                "flex items-center gap-2",
                isMobile && "w-full flex-col"
              )}
            >
              <Button
                size="sm"
                className={cn(isMobile && "w-full")}
                onClick={() => onCheckService(serviceDetailData?.service.id)}
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
        </Card>
        <Card>
          <CardHeader className="border-b">
            <CardTitle
              className={cn("text-lg font-bold", isMobile && "text-base px-4")}
            >
              Service Information
            </CardTitle>
          </CardHeader>
          <CardContent
            className={cn("flex flex-col gap-2 px-6", isMobile && "px-4")}
          >
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Status:
              </span>
              <span>
                <Badge>{serviceDetailData?.state.status}</Badge>
              </span>
            </div>
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Enabled:
              </span>
              <span>
                <ActivityIndicatorSVG
                  active={serviceDetailData?.service.is_enabled}
                  size={24}
                />
              </span>
            </div>
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Protocol:
              </span>
              <span className="font-medium">
                {serviceDetailData?.service.protocol}
              </span>
            </div>
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Total Checks:
              </span>
              <span className="font-medium">
                {serviceDetailData?.state.total_checks}
              </span>
            </div>
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Consecutive Success:
              </span>
              <span className="font-medium">
                {serviceDetailData?.state.consecutive_success}
              </span>
            </div>
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Consecutive Fails:
              </span>
              <span className="font-medium">
                {serviceDetailData?.state.consecutive_fails}
              </span>
            </div>
            <div className="flex items-center gap-2 justify-between">
              <span className="text-muted-foreground text-sm font-medium">
                Last Check:
              </span>
              <span className="font-medium">
                {new Date(serviceDetailData?.state.last_check).toLocaleString(
                  "ru-RU",
                  {
                    year: "numeric",
                    month: "numeric",
                    day: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit",
                  }
                )}
              </span>
            </div>
          </CardContent>
        </Card>
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
                <div className="flex items-center justify-between">
                  <span>
                    {new Date(incident.start_time).toLocaleString("ru-RU", {
                      year: "numeric",
                      month: "numeric",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                      second: "2-digit",
                    })}
                  </span>
                  <div className="flex items-center gap-2">
                    <span className="text-muted-foreground text-sm font-medium">
                      Resolved
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
                <span className="text-destructive text-sm font-medium">
                  {incident.error}
                </span>
                <div className="flex flex-col gap-2 incident-details">
                  <div className="flex items-center gap-2">
                    <strong>Start:</strong>{" "}
                    {new Date(incident.start_time).toLocaleString("ru-RU", {
                      year: "numeric",
                      month: "numeric",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                      second: "2-digit",
                    })}
                  </div>
                  <div className="flex items-center gap-2">
                    <strong>End:</strong>{" "}
                    {new Date(incident.end_time).toLocaleString("ru-RU", {
                      year: "numeric",
                      month: "numeric",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                      second: "2-digit",
                    })}
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
                <hr className="mt-4" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </ContentWrapper>
  );
};

export default ServiceDetail;
