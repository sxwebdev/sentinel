import { create } from "zustand";
import type { Service } from "../../../features/service/types/type";

interface ServiceTableStore {
  data: Service[] | null;
  deleteServiceId: string | null;
  updateServiceId: string | null;
  isOpenDropdownIdAction: string | null;
  isLoadingAllServices: boolean;
  filters: {
    search: string;
    page: number;
  };
  setData: (data: Service[] | null) => void;
  setSearch: (search: string) => void;
  setUpdateService: (updateService: Service | null) => void;
  setPage: (page: number) => void;
  setUpdateAllServices: (updateService: Service | null) => void;
  setIsLoadingAllServices: (isLoadingAllServices: boolean) => void;
  setDeleteServiceId: (deleteServiceId: string | null) => void;
  setUpdateServiceId: (updateServiceId: string | null) => void;
  setIsOpenDropdownIdAction: (isOpenDropdownIdAction: string | null) => void;
  deleteServiceInData: (deleteServiceId: string) => void;
  addServiceInData: (service: Service) => void;
}

const initialState = {
  data: null,
  deleteServiceId: null,
  updateServiceId: null,
  isOpenDropdownIdAction: null,
  isLoadingAllServices: false,
  filters: {
    search: "",
    page: 1,
  },
};

export const useServiceTableStore = create<ServiceTableStore>((set) => ({
  ...initialState,
  setData: (data) => set({ data }),
  setIsOpenDropdownIdAction: (isOpenDropdownIdAction) =>
    set({ isOpenDropdownIdAction }),
  setDeleteServiceId: (deleteServiceId) => set({ deleteServiceId }),
  setUpdateServiceId: (updateServiceId) => set({ updateServiceId }),
  setSearch: (search) => set({ filters: { ...initialState.filters, search } }),
  setPage: (page) => set({ filters: { ...initialState.filters, page } }),
  setUpdateService: (updateService) =>
    set((state) => {
      if (!updateService) return { data: state.data };
      const exists = state.data?.some(
        (ser) => ser.service.id === updateService.service.id,
      );
      return {
        data: exists
          ? state.data?.map((ser) =>
              ser.service.id === updateService.service.id ? updateService : ser,
            )
          : [...(state.data ?? []), updateService],
      };
    }),
  setIsLoadingAllServices: (isLoadingAllServices) =>
    set({ isLoadingAllServices }),
  setUpdateAllServices: (updateService) =>
    set((state) => {
      if (!updateService) return { data: state.data };
      const exists = state.data?.some(
        (ser) => ser.service.id === updateService.service.id,
      );
      return {
        data: exists
          ? state.data?.map((ser) =>
              ser.service.id === updateService.service.id ? updateService : ser,
            )
          : [...(state.data ?? []), updateService],
      };
    }),
  deleteServiceInData: (deleteServiceId) =>
    set((state) => {
      const exists = state.data?.some(
        (ser) => ser.service.id === deleteServiceId,
      );

      return {
        data: exists
          ? state.data?.filter((ser) => ser.service.id !== deleteServiceId)
          : state.data,
      };
    }),
  addServiceInData: (service) =>
    set((state) => {
      return { data: [...(state.data ?? []), service] };
    }),
}));
