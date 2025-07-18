import { create } from "zustand";
import type {
  Incident,
  Service,
  ServiceStats,
} from "../../../features/service/types/type";

interface ServiceDetailStore {
  deleteIncident: Incident | null;
  serviceDetailData: Service | null;
  resolveIncident: boolean;
  incidentsCount: number | null;
  incidentsData: Incident[] | null;
  filters: {
    page: number;
    pageSize: number;
  };
  serviceStatsData: ServiceStats | null;
  setFilters: (value: Partial<ServiceDetailStore["filters"]>) => void;
  setIncidentsCount: (incidentsCount: number) => void;
  setDeleteIncident: (deleteIncident: Incident | null) => void;
  setResolveIncident: (resolveIncident: boolean) => void;
  setServiceDetailData: (serviceDetailData: Service | null) => void;
  setIncidentsData: (incidentsData: Incident[] | null) => void;
  setServiceStatsData: (serviceStatsData: ServiceStats | null) => void;
  setUpdateServiceStatsData: (serviceStatsData: Service | null) => void;
}

const initialState = {
  deleteIncident: null,
  serviceDetailData: null,
  resolveIncident: false,
  incidentsData: null,
  incidentsCount: null,
  filters: {
    page: 1,
    pageSize: 10,
  },
  serviceStatsData: null,
};

export const useServiceDetailStore = create<ServiceDetailStore>((set) => ({
  ...initialState,
  setDeleteIncident: (deleteIncident) => set({ deleteIncident }),
  setIncidentsCount: (incidentsCount) => set({ incidentsCount }),
  setFilters: (filters) => set((state) => ({filters: {...state.filters, ...filters}})),
  setResolveIncident: (resolveIncident) => set({ resolveIncident }),
  setServiceDetailData: (serviceDetailData) => set({ serviceDetailData }),
  setIncidentsData: (incidentsData) => set({ incidentsData }),
  setServiceStatsData: (serviceStatsData) => set({ serviceStatsData }),
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
