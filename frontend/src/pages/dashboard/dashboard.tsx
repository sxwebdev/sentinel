import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
  Card,
  Progress,
} from "@/shared/components/ui";
import { useDashboardLogic } from "./hooks/useDashboardLogic";
import { InfoCardStats } from "@/entities/infoStatsCard/infoCardStats";
import { Loader } from "@/entities/loader/loader";
import type { GetDashboardStatsResult } from "@/shared/api/dashboard/dashboard";
import { getProtocolDisplayName } from "@/shared/lib/getProtocolDisplayName";
import { ServiceTable } from "../service/serviceTable";
import { useWsLogic } from "./hooks/useWsLogic";
import { ChartIncidentsStats } from "./incidents-stats";

const formatNumber = (value: number) => {
  return Intl.NumberFormat().format(value);
};

const infoKeysDashboard = [
  { key: "total_services", label: "Total services" },
  { key: "services_up", label: "Services up" },
  { key: "services_down", label: "Services down" },
  { key: "active_incidents", label: "Active incidents" },
  {
    key: "avg_response_time",
    label: "Average response time (ms)",
    valueFormatter: (value: string) => `${formatNumber(Number(value))}ms`,
  },
  {
    key: "total_checks",
    label: "Total checks",
    valueFormatter: (value: string) => `${formatNumber(Number(value))}`,
  },
  {
    key: "uptime_percentage",
    label: "Uptime",
    valueFormatter: (value: string) => `${Number(value).toFixed(1)}%`,
  },
  {
    key: "checks_per_minute",
    label: "Checks per minute",
    valueFormatter: (value: string) => `${formatNumber(Number(value))}`,
  },
];

const Dashboard = () => {
  const { dashboardInfo } = useDashboardLogic();
  useWsLogic();

  if (!dashboardInfo) return <Loader loaderPage />;

  return (
    <div className="flex flex-col gap-4 lg:gap-6">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-3 md:gap-6 lg:grid-cols-4">
        {infoKeysDashboard.map((item) => {
          const value =
            dashboardInfo[
              item.key as keyof GetDashboardStatsResult
            ]?.toString() || "-";

          return (
            <InfoCardStats
              key={item.key}
              title={item.label}
              value={item.valueFormatter ? item.valueFormatter(value) : value}
            />
          );
        })}
      </div>

      <ChartIncidentsStats />

      <div className="hidden">
        <Accordion type="multiple">
          <AccordionItem value="item-1" className="rounded-lg shadow-sm">
            <AccordionTrigger className="border-border flex cursor-pointer items-center justify-between border bg-white px-6 py-4 text-lg hover:no-underline">
              <h3 className="no-underline">Distribution by protocol</h3>
            </AccordionTrigger>
            <AccordionContent className="flex flex-col gap-4 rounded-b-lg bg-white px-6 py-4">
              {dashboardInfo?.protocols &&
              Object.entries(dashboardInfo.protocols).length > 0 ? (
                Object.entries(dashboardInfo.protocols).map(
                  ([protocol, count]) => {
                    const totalCount = Object.values(
                      dashboardInfo.protocols!,
                    ).reduce((a, b) => a + b, 0);
                    const percentage =
                      totalCount > 0
                        ? ((count / totalCount) * 100).toFixed(1)
                        : "0.0";

                    return (
                      <Card
                        key={protocol}
                        className="flex flex-col items-center justify-between gap-2 p-4 md:flex-row"
                      >
                        <h3 className="text-lg font-bold">
                          {getProtocolDisplayName(protocol)}
                        </h3>
                        <div className="flex flex-row items-center gap-2">
                          <p className="text-muted-foreground text-lg">
                            {count}
                          </p>
                          <Progress
                            value={Number(percentage)}
                            className="h-2 w-[100px]"
                            max={100}
                          />
                          <p className="text-muted-foreground">{percentage}%</p>
                        </div>
                      </Card>
                    );
                  },
                )
              ) : (
                <p className="text-muted-foreground text-center">
                  No protocols found
                </p>
              )}
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
      <div>
        <ServiceTable protocols={dashboardInfo.protocols ?? {}} />
      </div>
    </div>
  );
};

export default Dashboard;
