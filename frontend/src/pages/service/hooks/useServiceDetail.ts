import { socketUrl } from "@/shared/api/baseApi";
import { useEffect } from "react";
import { toast } from "sonner";
import { useServiceDetailStore } from "../store/useServiceDeteilStore";
import useWebSocket from "react-use-websocket";
import { getServices } from "@/shared/api/services/services";
import { getIncidents } from "@/shared/api/incidents/incidents";
import { getStatistics } from "@/shared/api/statistics/statistics";

export const useServiceDetail = (serviceID: string) => {
  const {
    deleteIncident,
    serviceDetailData,
    incidentsData,
    serviceStatsData,
    resolveIncident,
    filters,
    setDeleteIncident,
    setServiceDetailData,
    setIncidentsData,
    setServiceStatsData,
    setFilters,
    setUpdateServiceStatsData,
    setResolveIncident,
  } = useServiceDetailStore();

  const { postServicesIdCheck, getServicesId } = getServices();
  const { getServicesIdStats } = getStatistics();
  const {
    getServicesIdIncidents,
    deleteServicesIdIncidentsIncidentId,
    postServicesIdResolve,
  } = getIncidents();

  // Get service
  const getServiceDetail = async () => {
    return await getServicesId(serviceID ?? "").then((res) => {
      setServiceDetailData(res);
    });
    // .catch(() => {
    //   navigate(ROUTES.NOT_FOUND);
    // });
  };

  // Get service stats
  const getServiceStats = async () => {
    return await getServicesIdStats(serviceID ?? "").then((res) => {
      setServiceStatsData(res);
    });
    // .catch(() => {
    //   navigate(ROUTES.NOT_FOUND);
    // });
  };

  //Check service
  const onCheckService = async (id: string) => {
    await postServicesIdCheck(id)
      .then(() => {
        getServiceDetail();
        getIncidents();
        getServiceStats();
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
  };

  // Get all incidents
  const getAllIncidents = async () => {
    return await getServicesIdIncidents(serviceID ?? "", filters).then(
      (res) => {
        setIncidentsData(res);
      }
    );
  };

  // Delete incident
  const onDeleteIncident = async (incidentId: string) => {
    await deleteServicesIdIncidentsIncidentId(serviceID ?? "", incidentId)
      .then(() => {
        getServiceDetail();
        getAllIncidents();
        getServiceStats();
        toast.success("Incident deleted");
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      })
      .finally(() => {
        setDeleteIncident(null);
      });
  };

  // Resolve incident
  const onResolveIncident = async () => {
    await postServicesIdResolve(serviceID ?? "")
      .then(() => {
        getServiceDetail();
        getAllIncidents();
        getServiceStats();
        toast.success("Incident resolved");
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      })
      .finally(() => {
        setResolveIncident(false);
      });
  };

  // WebSocket connection to update service stats
  const { lastMessage } = useWebSocket(socketUrl, {
    shouldReconnect: () => true,
  });

  useEffect(() => {
    if (!lastMessage) return;
    const data = JSON.parse(lastMessage.data);
    switch (data.type) {
      case "service_updated_state":
        if (data.data.id === serviceID) {
          setUpdateServiceStatsData(data.data);
        }
        break;
    }
  }, [lastMessage]);

  // const navigate = useNavigate();

  useEffect(() => {
    getServiceDetail();
    getServiceStats();
    getAllIncidents();
    return () => {
      setServiceDetailData(null);
      setIncidentsData(null);
      setServiceStatsData(null);
    };
  }, [serviceID]);

  useEffect(() => {
    getAllIncidents();
  }, [filters, serviceID]);

  return {
    deleteIncident,
    serviceDetailData,
    incidentsData,
    serviceStatsData,
    resolveIncident,
    filters,
    onCheckService,
    onDeleteIncident,
    onResolveIncident,
    setFilters,
    setDeleteIncident,
    setResolveIncident,
  };
};
