import $api, { socketUrl } from "@/shared/api/baseApi";
import { useEffect, useMemo } from "react";
import useWebSocket from "react-use-websocket";
import { useShallow } from "zustand/react/shallow";
import { useServiceTableStore } from "@/pages/service/store/useServiceTableStore";
import { useDashboardStore } from "../store/useDashboardStore";

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
  const {
    deleteServiceInData,
    setUpdateService,
    setUpdateAllServices,
    addServiceInData,
  } = useServiceTableStore(
    useShallow((s) => ({
      deleteServiceInData: s.deleteServiceInData,
      setUpdateService: s.setUpdateService,
      setUpdateAllServices: s.setUpdateAllServices,
      addServiceInData: s.addServiceInData,
    })),
  );

  const { dashboardInfo, setDashboardInfo } = useDashboardStore(
    useShallow((s) => ({
      dashboardInfo: s.dashboardInfo,
      setDashboardInfo: s.setDashboardInfo,
    })),
  );

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

  const { lastMessage } = useWebSocket(socketUrl, {
    shouldReconnect: () => true,
  });

  useEffect(() => {
    if (!lastMessage) return;
    const data = JSON.parse(lastMessage.data);
    switch (data.type) {
      case "stats_update":
        setDashboardInfo(data.data);
        break;
      case "service_deleted":
        deleteServiceInData(data.data.service.id);
        break;
      case "service_updated":
        setUpdateService(data.data);
        break;
      case "service_created":
        addServiceInData(data.data);
        break;
      case "service_updated_state":
        setUpdateAllServices(data.data);
        break;
    }
  }, [lastMessage]);

  const infoKeysDashboard = useMemo(
    () => [
      { key: "total_services", label: "Total Services" },
      { key: "services_up", label: "Services Up" },
      { key: "services_down", label: "Services Down" },
      { key: "active_incidents", label: "Active Incidents" },
      { key: "avg_response_time", label: "Average Response Time (ms)" },
      { key: "total_checks", label: "Total Checks" },
      { key: "uptime_percentage", label: "Uptime Percentage" },
      { key: "checks_per_minute", label: "Checks Per Minute" },
    ],
    [],
  );

  return {
    dashboardInfo,
    infoKeysDashboard,
    onRefreshDashboard,
  };
};
