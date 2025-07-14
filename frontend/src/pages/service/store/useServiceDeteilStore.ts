import {create} from "zustand";
import type {
  Incident,
  Service,
  ServiceStats,
} from "../../../features/service/types/type";

interface ServiceDetailStore {
  deleteIncident: Incident | null;
  serviceDetailData: Service | null;
  resolveIncident: boolean;
  incidentsData: Incident[] | null;
  serviceStatsData: ServiceStats | null;
  setDeleteIncident: (deleteIncident: Incident | null) => void;
  setResolveIncident: (resolveIncident: boolean) => void;
  setServiceDetailData: (serviceDetailData: Service | null) => void;
  setIncidentsData: (incidentsData: Incident[] | null) => void;
  setServiceStatsData: (serviceStatsData: ServiceStats | null) => void;
}

const initialState = {
  deleteIncident: null,
  serviceDetailData: null,
  resolveIncident: false,
  incidentsData: null,
  serviceStatsData: null,
};

export const useServiceDetailStore = create<ServiceDetailStore>((set) => ({
  ...initialState,
  setDeleteIncident: (deleteIncident) => set({deleteIncident}),
  setResolveIncident: (resolveIncident) => set({resolveIncident}),
  setServiceDetailData: (serviceDetailData) => set({serviceDetailData}),
  setIncidentsData: (incidentsData) => set({incidentsData}),
  setServiceStatsData: (serviceStatsData) => set({serviceStatsData}),
}));
