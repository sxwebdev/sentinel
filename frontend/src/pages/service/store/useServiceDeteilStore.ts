import type {
  StorecmnFindResponseWithCountStorageIncident,
  GetServicesIdIncidentsParams,
  StorageIncident,
  WebServiceDTO,
  ServiceStats,
} from "@/shared/types/model";
import { create } from "zustand";

interface ServiceDetailStore {
  deleteIncident: StorageIncident | null;
  serviceDetailData: WebServiceDTO | null;
  resolveIncident: boolean;
  incidentsData: StorecmnFindResponseWithCountStorageIncident | null;
  filters: GetServicesIdIncidentsParams;
  serviceStatsData: ServiceStats | null;
  setFilters: (value: Partial<ServiceDetailStore["filters"]>) => void;
  setDeleteIncident: (deleteIncident: StorageIncident | null) => void;
  setResolveIncident: (resolveIncident: boolean) => void;
  setServiceDetailData: (serviceDetailData: WebServiceDTO | null) => void;
  setIncidentsData: (
    incidentsData: StorecmnFindResponseWithCountStorageIncident | null,
  ) => void;
  setServiceStatsData: (serviceStatsData: ServiceStats | null) => void;
  setUpdateServiceStatsData: (serviceStatsData: ServiceStats | null) => void;
}

const initialState = {
  deleteIncident: null,
  serviceDetailData: null,
  resolveIncident: false,
  incidentsData: null,

  filters: {
    page: 1,
    page_size: 10,
  },
  serviceStatsData: null,
};

export const useServiceDetailStore = create<ServiceDetailStore>((set) => ({
  ...initialState,
  setDeleteIncident: (deleteIncident) => set({ deleteIncident }),
  setFilters: (filters) =>
    set((state) => ({ filters: { ...state.filters, ...filters } })),
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
