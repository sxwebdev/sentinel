import {socketUrl} from "@/shared/api/baseApi";
import {useEffect} from "react";
import {useNavigate, useParams} from "react-router";
import {toast} from "sonner";
import {useServiceDetailStore} from "../store/useServiceDeteilStore";
import useWebSocket from "react-use-websocket";
import {ROUTES} from "@/app/routes/constants";
import { getServices } from "@/shared/api/services/services";
import { getIncidents } from "@/shared/api/incidents/incidents";
import { getStatistics } from "@/shared/api/statistics/statistics";

export const useServiceDetail = () => {
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
  const {id} = useParams();

  const {postServicesIdCheck, getServicesId} = getServices();
  const {getServicesIdStats} = getStatistics();
  const {
    getServicesIdIncidents,
    deleteServicesIdIncidentsIncidentId,
    postServicesIdResolve,
  } = getIncidents();

  // Get service
  const getServiceDetail = async () => {
    return await getServicesId(id ?? "")
      .then((res) => {
        setServiceDetailData(res);
      })
      .catch(() => {
        navigate(ROUTES.NOT_FOUND);
      });
  };

  // Get service stats
  const getServiceStats = async () => {
    return await getServicesIdStats(id ?? "")
      .then((res) => {
        setServiceStatsData(res);
      })
      .catch(() => {
        navigate(ROUTES.NOT_FOUND);
      });
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
    return await getServicesIdIncidents(id ?? "", filters).then((res) => {
      setIncidentsData(res);
    });
  };

  // Delete incident
  const onDeleteIncident = async (incidentId: string) => {
    await deleteServicesIdIncidentsIncidentId(id ?? "", incidentId)
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
    await postServicesIdResolve(id ?? "")
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
  const {lastMessage} = useWebSocket(socketUrl, {
    shouldReconnect: () => true,
  });

  useEffect(() => {
    if (!lastMessage) return;
    const data = JSON.parse(lastMessage.data);
    switch (data.type) {
      case "service_updated_state":
        if (data.data.id === id) {
          setUpdateServiceStatsData(data.data);
        }
        break;
    }
  }, [lastMessage]);

  const navigate = useNavigate();

  useEffect(() => {
    getServiceDetail();
    getServiceStats();
    getAllIncidents();
    return () => {
      setServiceDetailData(null);
      setIncidentsData(null);
      setServiceStatsData(null);
    };
  }, [id]);

  useEffect(() => {
    getAllIncidents();
  }, [filters, id]);

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
