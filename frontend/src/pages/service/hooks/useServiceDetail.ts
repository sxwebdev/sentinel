import $api, { socketUrl } from "@/shared/api/baseApi";
import { useEffect } from "react";
import { useParams } from "react-router";
import { useServiceApi } from "./useServiceApi";
import { toast } from "sonner";
import { useServiceDetailStore } from "../store/useServiceDeteilStore";
import useWebSocket from "react-use-websocket";

export const useServiceDetail = () => {
  const {
    deleteIncident,
    serviceDetailData,
    incidentsData,
    serviceStatsData,
    resolveIncident,
    incidentsCount,
    filters,
    setDeleteIncident,
    setServiceDetailData,
    setIncidentsData,
    setIncidentsCount,
    setServiceStatsData,
    setFilters,
    setUpdateServiceStatsData,
    setResolveIncident,
  } = useServiceDetailStore();
  const { id } = useParams();
  const { onCheckService: onCheckServiceApi } = useServiceApi();

  const onCheckService = async (id: string) => {
    await onCheckServiceApi(id)
      .then(() => {
        getServiceDetail();
        getIncidents();
        getServiceStats();
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
  };

  const { lastMessage } = useWebSocket(socketUrl, {
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
    await $api
      .delete(`/services/${id}/incidents/${incidentId}`)
      .then(() => {
        getServiceDetail();
        getIncidents();
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

  const onResolveIncident = async () => {
    await $api
      .post(`/services/${id}/resolve`)
      .then(() => {
        getServiceDetail();
        getIncidents();
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

  const getServiceDetail = async () => {
    const res = await $api.get(`/services/${id}`);
    setServiceDetailData(res.data);
  };

  const getIncidents = async () => {
    const res = await $api.get(`/services/${id}/incidents`, {
      params: {
        page: filters.page,
        pageSize: filters.pageSize,
      },
    });
    setIncidentsData(res.data.items);
    setIncidentsCount(res.data.count);
  };

  const getServiceStats = async () => {
    const res = await $api.get(`/services/${id}/stats`);
    setServiceStatsData(res.data);
  };

  useEffect(() => {
    getServiceDetail();
    getIncidents();
    getServiceStats();
    return () => {
      setServiceDetailData(null);
      setIncidentsData(null);
      setServiceStatsData(null);
    };
  }, [id]);

  return {
    deleteIncident,
    serviceDetailData,
    incidentsData,
    serviceStatsData,
    incidentsCount,
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
