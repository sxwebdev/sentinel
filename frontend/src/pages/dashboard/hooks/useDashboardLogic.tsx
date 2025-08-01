import { socketUrl } from "@/shared/api/baseApi";
import { useEffect } from "react";
import useWebSocket from "react-use-websocket";
import { useShallow } from "zustand/react/shallow";
import { useServiceTableStore } from "@/pages/service/store/useServiceTableStore";
import { useDashboardStore } from "../store/useDashboardStore";

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
    }))
  );

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
  }, []);

  const { lastMessage } = useWebSocket(socketUrl, {
    shouldReconnect: () => true,
  });

  useEffect(() => {
    if (!lastMessage) return;
    const data = JSON.parse(lastMessage.data);
    switch (data.type) {
      case "stats_update":
        setStats(data.data);
        break;
      case "service_deleted":
        deleteServiceInData(data.data.id);
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

  return {
    dashboardInfo,
    onRefreshDashboard,
  };
};
