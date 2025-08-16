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
import { CircleAlertIcon, PlayIcon } from "lucide-react";
import { useIsMobile } from "@/shared/hooks/useIsMobile";
import { cn } from "@/shared/lib/utils";
import { ActivityIndicatorSVG } from "@/entities/ActivityIndicatorSVG/ActivityIndicatorSVG";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/shared/components/ui/alert";
import type { WebServiceDTO } from "@/shared/types/model";

interface ServiceOverviewProps {
  serviceDetailData: WebServiceDTO;
  onCheckService: (serviceId: string) => void;
  setResolveIncident: (value: boolean) => void;
}

export const ServiceOverview = ({
  serviceDetailData,
  onCheckService,
}: ServiceOverviewProps) => {
  const isMobile = useIsMobile();

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-3.5 md:gap-5 flex-col md:flex-row">
            <Badge
              className={cn(
                "text-xs md:text-sm font-semibold",
                serviceDetailData?.status === "up" &&
                  "bg-emerald-100 text-emerald-600",
                serviceDetailData?.status === "down" &&
                  "bg-rose-100 text-rose-600",
                serviceDetailData?.status === "unknown" &&
                  "bg-yellow-100 text-yellow-600"
              )}
            >
              {serviceDetailData?.status?.toLocaleUpperCase() ?? ""}
            </Badge>

            <div className="flex flex-col w-full">
              <h3 className="text-base md:text-lg font-bold">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <ActivityIndicatorSVG
                        active={serviceDetailData?.is_enabled}
                        size={24}
                      />
                    </TooltipTrigger>
                    <TooltipContent className="dark" showArrow>
                      <p>
                        {serviceDetailData?.is_enabled
                          ? "Service enabled"
                          : "Service disabled"}
                      </p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
                <span className="ml-2">{serviceDetailData?.name}</span>
              </h3>
              <div className="mt-2">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <Badge variant={"secondary"} className="text-sm">
                        {serviceDetailData?.protocol}
                      </Badge>
                    </TooltipTrigger>
                    <TooltipContent className="dark" showArrow>
                      Protocol
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <Badge variant={"secondary"} className="ml-3 text-sm">
                        {new Date(
                          serviceDetailData?.last_check ?? ""
                        ).toLocaleString()}
                      </Badge>
                    </TooltipTrigger>
                    <TooltipContent className="dark" showArrow>
                      Last Check
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            </div>

            <Button
              size="sm"
              className={cn("ml-auto", isMobile && "w-full")}
              disabled={!serviceDetailData?.is_enabled}
              onClick={() => onCheckService(serviceDetailData?.id ?? "")}
            >
              <PlayIcon />
              Trigger Check
            </Button>
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
