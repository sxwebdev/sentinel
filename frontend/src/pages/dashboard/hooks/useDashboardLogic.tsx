import {socketUrl} from "@/shared/api/baseApi";
import {useEffect, useMemo} from "react";
import useWebSocket from "react-use-websocket";
import {useShallow} from "zustand/react/shallow";
import {useServiceTableStore} from "@/pages/service/store/useServiceTableStore";
import {useDashboardStore} from "../store/useDashboardStore";
import {getDashboard} from "@/shared/api/dashboard/dashboard";

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
  
  const {getDashboardStats} = getDashboard();

  const {dashboardInfo, setDashboardInfo} = useDashboardStore(
    useShallow((s) => ({
      dashboardInfo: s.dashboardInfo,
      setDashboardInfo: s.setDashboardInfo,
    }))
  );



  const onRefreshDashboard = async () => {
    await getDashboardStats();
  };

  useEffect(() => {
    getDashboardStats().then((res) => {
      setDashboardInfo(res);
    });

    return () => {
      setDashboardInfo(null);
    };
  }, []);

  const {lastMessage} = useWebSocket(socketUrl, {
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

  const infoKeysDashboard = useMemo(
    () => [
      {key: "total_services", label: "Total services"},
      {key: "services_up", label: "Services up"},
      {key: "services_down", label: "Services down"},
      {key: "active_incidents", label: "Active incidents"},
      {key: "avg_response_time", label: "Average response time (ms)"},
      {key: "total_checks", label: "Total checks"},
      {key: "uptime_percentage", label: "Uptime"},
      {key: "checks_per_minute", label: "Checks per minute"},
    ],
    []
  );

  return {
    dashboardInfo,
    infoKeysDashboard,
    onRefreshDashboard,
  };
};
