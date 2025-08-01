import { useDashboardStore } from "@/pages/dashboard/store/useDashboardStore";
import { useEffect } from "react";
import { socketUrl } from "@/shared/api/baseApi";
import useWebSocket from "react-use-websocket";
import { useServiceTableStore } from "@/pages/service/store/useServiceTableStore";
import { useShallow } from "zustand/react/shallow";

export const useWsLogic = () => {
  const { setDashboardInfo } = useDashboardStore();
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
};
