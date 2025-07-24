import type {
  DbutilsFindResponseWithCountWebIncident,
  GetServicesIdIncidentsParams,
  WebServiceDTO,
} from "@/shared/types/model";
import {create} from "zustand";

interface ServiceDetailStore {
  deleteIncident: Incident | null;
  serviceDetailData: Service | null;
  resolveIncident: boolean;
  incidentsData: DbutilsFindResponseWithCountWebIncident | null;
  filters: GetServicesIdIncidentsParams;
  serviceStatsData: WebServiceDTO | null;
  setFilters: (value: Partial<ServiceDetailStore["filters"]>) => void;
  setDeleteIncident: (deleteIncident: Incident | null) => void;
  setResolveIncident: (resolveIncident: boolean) => void;
  setServiceDetailData: (serviceDetailData: WebServiceDTO | null) => void;
  setIncidentsData: (
    incidentsData: DbutilsFindResponseWithCountWebIncident | null
  ) => void;
  setServiceStatsData: (serviceStatsData: WebServiceDTO | null) => void;
  setUpdateServiceStatsData: (serviceStatsData: WebServiceDTO | null) => void;
}

const initialState = {
  deleteIncident: null,
  serviceDetailData: null,
  resolveIncident: false,
  incidentsData: null,

  filters: {
    page: 1,
    pageSize: 10,
  },
  serviceStatsData: null,
};

export const useServiceDetailStore = create<ServiceDetailStore>((set) => ({
  ...initialState,
  setDeleteIncident: (deleteIncident) => set({deleteIncident}),
  setFilters: (filters) =>
    set((state) => ({filters: {...state.filters, ...filters}})),
  setResolveIncident: (resolveIncident) => set({resolveIncident}),
  setServiceDetailData: (serviceDetailData) => set({serviceDetailData}),
  setIncidentsData: (incidentsData) => set({incidentsData}),
  setServiceStatsData: (serviceStatsData) => set({serviceStatsData}),
  setUpdateServiceStatsData: (serviceStatsData) =>
    set((store) => {
      if (!serviceStatsData) return store;

      return {
        serviceDetailData: {
          ...serviceStatsData,
        },
      };
    }),
}));
