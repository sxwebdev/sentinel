import { useEffect, useMemo } from "react";
import { useShallow } from "zustand/react/shallow";
import { useDashboardStore } from "../store/useDashboardStore";
import { getDashboard } from "@/shared/api/dashboard/dashboard";

export const useDashboardLogic = () => {
  const { getDashboardStats } = getDashboard();

  const { dashboardInfo, setDashboardInfo } = useDashboardStore(
    useShallow((s) => ({
      dashboardInfo: s.dashboardInfo,
      setDashboardInfo: s.setDashboardInfo,
    }))
  );

  useEffect(() => {
    getDashboardStats().then((res) => {
      setDashboardInfo(res);
    });
  }, []);

  const onRefreshDashboard = async () => {
    await getDashboardStats();
  };

  const infoKeysDashboard = useMemo(
    () => [
      { key: "total_services", label: "Total services" },
      { key: "services_up", label: "Services up" },
      { key: "services_down", label: "Services down" },
      { key: "active_incidents", label: "Active incidents" },
      { key: "avg_response_time", label: "Average response time (ms)" },
      { key: "total_checks", label: "Total checks" },
      { key: "uptime_percentage", label: "Uptime" },
      { key: "checks_per_minute", label: "Checks per minute" },
    ],
    []
  );

  return {
    dashboardInfo,
    infoKeysDashboard,
    onRefreshDashboard,
  };
};
