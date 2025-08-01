import { InfoCardStats } from "@/entities/infoStatsCard/infoCardStats";
import type { WebServiceDTO, WebServiceStats } from "@/shared/types/model";

interface ServiceStatsProps {
  serviceDetailData: WebServiceDTO;
  serviceStatsData: WebServiceStats;
}

export const ServiceStats = ({
  serviceDetailData,
  serviceStatsData,
}: ServiceStatsProps) => {
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
      value: `${((serviceStatsData?.avg_response_time ?? 0) / 1000000).toFixed(1)} ms`,
      key: "avg_response_time",
      description: "Avg Response Time",
    },
    {
      value: `${(serviceStatsData?.uptime_percentage ?? 0).toFixed(1)}%`,
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
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {cardStats.map((stat) => (
        <InfoCardStats
          key={stat.key}
          title={stat.description}
          value={stat?.value?.toString() ?? "0"}
        />
      ))}
    </div>
  );
};
