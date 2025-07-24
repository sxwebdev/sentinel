import $api, {socketUrl} from "@/shared/api/baseApi";
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
  const { id } = useParams();
  
  const { postServicesIdCheck, getServicesId } = getServices();
  const { getServicesIdStats } = getStatistics();
  const {getServicesIdIncidents, deleteServicesIdIncidentsIncidentId, postServicesIdResolve} =
    getIncidents();



  const onCheckService = async (id: string) => {
    await postServicesIdCheck(id)
      .then(() => {
        getServiceDetail();
        getIncidents();
        getServicesIdStats(id);
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
  };

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

  const onDeleteIncident = async (incidentId: string) => {
    await deleteServicesIdIncidentsIncidentId(id ?? "", incidentId)
      .then(() => {
        getServicesId(id ?? "");
        getServicesIdIncidents(id ?? "");
        getServicesIdStats(id ?? "");
        toast.success("Incident deleted");
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      })
      .finally(() => {
        setDeleteIncident(null);
      });
  };

  const onResolveIncident = async () => {
    await postServicesIdResolve(id ?? "")
      .then(() => {
        getServicesId(id ?? "");
        getServicesIdIncidents(id ?? "");
        getServicesIdStats(id ?? "");
        toast.success("Incident resolved");
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      })
      .finally(() => {
        setResolveIncident(false);
      });
  };

  const navigate = useNavigate();

  const getServiceDetail = async () => {
    return await getServicesId(id ?? "").then((res) => {
      setServiceDetailData(res);
    }).catch(() => {
      navigate(ROUTES.NOT_FOUND);
    });

  };


  const getServiceStats = async () => {
    return await getServicesIdStats(id ?? "").then((res) => {
      setServiceStatsData(res);
    }).catch(() => {
      navigate(ROUTES.NOT_FOUND);
    });
  };

  useEffect(() => {
    getServiceDetail();
    getServiceStats();
    return () => {
      setServiceDetailData(null);
      setIncidentsData(null);
      setServiceStatsData(null);
    };
  }, [id]);

  useEffect(() => {
    getIncidents();
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
