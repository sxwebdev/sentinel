import type {
  StorecmnFindResponseWithCountModelsIncident,
  GetServicesIdIncidentsParams,
  ModelsIncident,
  WebServiceDTO,
  ServiceStats,
} from "@/shared/types/model";
import { create } from "zustand";

interface ServiceDetailStore {
  deleteIncident: ModelsIncident | null;
  serviceDetailData: WebServiceDTO | null;
  incidentsData: StorecmnFindResponseWithCountModelsIncident | null;
  filters: GetServicesIdIncidentsParams;
  serviceStatsData: ServiceStats | null;
  setFilters: (value: Partial<ServiceDetailStore["filters"]>) => void;
  setDeleteIncident: (deleteIncident: ModelsIncident | null) => void;
  setServiceDetailData: (serviceDetailData: WebServiceDTO | null) => void;
  setIncidentsData: (
    incidentsData: StorecmnFindResponseWithCountModelsIncident | null,
  ) => void;
  setServiceStatsData: (serviceStatsData: ServiceStats | null) => void;
  setUpdateServiceStatsData: (serviceStatsData: ServiceStats | null) => void;
}

const initialState = {
  deleteIncident: null,
  serviceDetailData: null,
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
