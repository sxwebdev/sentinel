import { useEffect } from "react";
import { useShallow } from "zustand/react/shallow";
import { useDashboardStore } from "../store/useDashboardStore";

export const useDashboardLogic = () => {
  const { dashboardInfo, setStats, loadStats } = useDashboardStore(
    useShallow((s) => ({
      dashboardInfo: s.dashboardInfo,
      setStats: s.setStats,
      loadStats: s.loadStats,
    }))
  );

  const onRefreshDashboard = async () => {
    await loadStats();
  };

  useEffect(() => {
    loadStats();

    return () => {
      setStats(null);
    };
  }, [loadStats, setStats]);

  return {
    dashboardInfo,
    onRefreshDashboard,
  };
};
