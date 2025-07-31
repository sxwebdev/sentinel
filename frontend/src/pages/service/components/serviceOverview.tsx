import {
  Badge,
  Button,
  Card,
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
} from "lucide-react";
import { useIsMobile } from "@/shared/hooks/useIsMobile";
import { cn } from "@/shared/lib/utils";
import { ActivityIndicatorSVG } from "@/entities/ActivityIndicatorSVG/ActivityIndicatorSVG";
import { Link } from "react-router";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/shared/components/ui/alert";

interface ServiceOverviewProps {
  serviceDetailData: {
    id: string;
    name: string;
    protocol: string;
    last_check: string;
    status: string;
    is_enabled: boolean;
    last_error?: string;
  };
  onCheckService: (serviceId: string) => void;
  setResolveIncident: (value: boolean) => void;
}

export const ServiceOverview = ({
  serviceDetailData,
  onCheckService,
  setResolveIncident,
}: ServiceOverviewProps) => {
  const isMobile = useIsMobile();

  return (
    <>
      <header className="flex flex-col py-3 md:flex-row gap-3 justify-between items-center">
        <Link to={"/"}>
          <Button className="group" variant="ghost" size="sm">
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
    </>
  );
};
