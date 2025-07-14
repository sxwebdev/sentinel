import $api from "@/shared/api/baseApi";
import {useEffect, useState} from "react";

export interface DashboardInfo {
  total_services: number;
  services_up: number;
  services_down: number;
  services_unknown: number;
  protocols: Record<string, number>;
  recent_incidents: number;
  active_incidents: number;
  avg_response_time: number;
  total_checks: number;
  uptime_percentage: number;
  last_check_time: string;
  checks_per_minute: number;
}

export const useDashboardLogic = () => {
  const [dashboardInfo, setDashboardInfo] = useState<DashboardInfo | null>(
    null
  );

  const infoKeysDashboard = [
    {
      key: "total_services",
      label: "Total Services",
    },
    {
      key: "services_up",
      label: "Services Up",
    },
    {
      key: "services_down",
      label: "Services Down",
    },
    {
      key: "active_incidents",
      label: "Active Incidents",
    },
    {
      key: "avg_response_time",
      label: "Average Response Time (ms)",
    },
    {
      key: "total_checks",
      label: "Total Checks",
    },
    {
      key: "uptime_percentage",
      label: "Uptime Percentage",
    },
    {
      key: "checks_per_minute",
      label: "Checks Per Minute",
    },
  ];

  const getDashboardInfo = async () => {
    const res = await $api.get("/dashboard/stats");
    setDashboardInfo(res.data);
  };

  const onRefreshDashboard = () => {
    getDashboardInfo();
  };

  useEffect(() => {
    getDashboardInfo();

    return () => {
      setDashboardInfo(null);
    };
  }, []);

  return {
    dashboardInfo,
    infoKeysDashboard,
    onRefreshDashboard,
  };
};
