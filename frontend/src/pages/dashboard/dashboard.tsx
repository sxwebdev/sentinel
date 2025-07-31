import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
  Button,
  Card,
  Progress,
} from "@/shared/components/ui";
import ContentWrapper from "@/widgets/wrappers/contentWrapper";
import {RefreshCcwIcon} from "lucide-react";
import {useDashboardLogic} from "./hooks/useDashboardLogic";
import {InfoCardStats} from "@/entities/infoStatsCard/infoCardStats";
import ServiceCreate from "../service/serviceCreate";
import {Loader} from "@/entities/loader/loader";
import type {GetDashboardStatsResult} from "@/shared/api/dashboard/dashboard";
import {getProtocolDisplayName} from "@/shared/lib/getProtocolDisplayName";
import {ServiceTable} from "../service/serviceTable";
import {ApiInfo} from "@/features/apiInfo/apiInfo";
import { UpdateVersion } from "@/features/apiInfo/updateVersion";

const Dashboard = () => {
  const {apiInfo, infoKeysDashboard, dashboardInfo, onRefreshDashboard} =
    useDashboardLogic();

  if (!dashboardInfo) return <Loader loaderPage />;

  return (
    <ContentWrapper>
      <div className="flex flex-col gap-4 lg:gap-6">
        <header className="flex flex-col py-3 md:flex-row gap-3 justify-between items-center">
          <div className="flex flex-row gap-3 items-center">
            <h1 className="text-lg md:text-2xl font-bold">
              Sentinel Dashboard
            </h1>
            {apiInfo?.available_update && <UpdateVersion apiInfo={apiInfo} />}
          </div>
          <div className="flex flex-col md:flex-row">
            <Button
              size="sm"
              className="mb-3 md:mb-0 md:mr-3"
              onClick={onRefreshDashboard}
            >
              <RefreshCcwIcon />
              Refresh
            </Button>
            <ServiceCreate />
          </div>
        </header>

        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 md:gap-6">
          {infoKeysDashboard.map((item) => {
            const value =
              item.key === "uptime_percentage"
                ? Number(
                    dashboardInfo[item.key as keyof GetDashboardStatsResult]
                  ).toFixed(1) + "%"
                : item.key === "avg_response_time"
                  ? dashboardInfo[
                      item.key as keyof GetDashboardStatsResult
                    ]?.toString() + "ms"
                  : (dashboardInfo[
                      item.key as keyof GetDashboardStatsResult
                    ]?.toString() ?? "0");

            return (
              <InfoCardStats key={item.key} title={item.label} value={value} />
            );
          })}
        </div>
        <div>
          <Accordion type="multiple">
            <AccordionItem value="item-1" className="shadow-sm rounded-lg">
              <AccordionTrigger className="bg-white flex justify-between items-center border hover:no-underline border-border cursor-pointer text-lg  py-4 px-6">
                <h3 className="no-underline">Distribution by protocol</h3>
              </AccordionTrigger>
              <AccordionContent className="px-6 py-4 bg-white rounded-b-lg flex flex-col gap-4">
                {dashboardInfo?.protocols &&
                Object.entries(dashboardInfo.protocols).length > 0 ? (
                  Object.entries(dashboardInfo.protocols).map(
                    ([protocol, count]) => {
                      const totalCount = Object.values(
                        dashboardInfo.protocols!
                      ).reduce((a, b) => a + b, 0);
                      const percentage =
                        totalCount > 0
                          ? ((count / totalCount) * 100).toFixed(1)
                          : "0.0";

                      return (
                        <Card
                          key={protocol}
                          className="p-4 flex md:flex-row justify-between items-center flex-col gap-2"
                        >
                          <h3 className="text-lg font-bold">
                            {getProtocolDisplayName(protocol)}
                          </h3>
                          <div className="flex flex-row gap-2 items-center">
                            <p className="text-lg text-muted-foreground">
                              {count}
                            </p>
                            <Progress
                              value={Number(percentage)}
                              className="w-[100px] h-2"
                              max={100}
                            />
                            <p className=" text-muted-foreground">
                              {percentage}%
                            </p>
                          </div>
                        </Card>
                      );
                    }
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
        <div className="flex flex-col gap-4 w-fit">
          <ApiInfo apiInfo={apiInfo} />
        </div>
      </div>
    </ContentWrapper>
  );
};

export default Dashboard;
