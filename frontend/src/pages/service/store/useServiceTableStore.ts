import {create} from "zustand";
import type {Service} from "../../../features/service/types/type";

interface ServiceTableStore {
  data: Service[] | null;
  deleteServiceId: string | null;
  updateServiceId: string | null;
  servicesCount: number | null;
  isOpenDropdownIdAction: string | null;
  isLoadingAllServices: boolean;
  allTags: string[] | null;
  countAllTags: Record<string, number> | null;
  filters: {
    search: string | undefined;
    page: number;
    pageSize: number;
    tags: string[] | undefined;
    protocol: string | undefined;
    status: string | undefined;
  };
  setData: (data: Service[] | null) => void;
  setFilters: (value: Partial<ServiceTableStore["filters"]>) => void;
  setServicesCount: (servicesCount: number) => void;
  setUpdateService: (updateService: Service | null) => void;
  setPage: (page: number) => void;
  setAllTags: (allTags: string[]) => void;
  setCountAllTags: (countAllTags: Record<string, number>) => void;
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
  allTags: null,
  countAllTags: null,
  servicesCount: null,
  isOpenDropdownIdAction: null,
  isLoadingAllServices: false,
  filters: {
    search: undefined,
    page: 1,
    pageSize: 10,
    tags: undefined,
    protocol: undefined,
    status: undefined,
  },
};

export const useServiceTableStore = create<ServiceTableStore>((set) => ({
  ...initialState,
  setData: (data) => set({data}),
  setIsOpenDropdownIdAction: (isOpenDropdownIdAction) =>
    set({isOpenDropdownIdAction}),
  setAllTags: (allTags) => set({allTags}),
  setCountAllTags: (countAllTags) => set({countAllTags}),
  setDeleteServiceId: (deleteServiceId) => set({deleteServiceId}),
  setFilters: (value) => set((state) => ({filters: {...state.filters, ...value}})),
  setUpdateServiceId: (updateServiceId) => set({updateServiceId}),
  setPage: (page) => set({filters: {...initialState.filters, page}}),
  setServicesCount: (servicesCount) => set({servicesCount}),
  setUpdateService: (updateService) =>
    set((state) => {
      if (!updateService) return {data: state.data};
      const exists = state.data?.some(
        (ser) => ser?.id === updateService?.id
      );
      return {
        data: exists
          ? state.data?.map((ser) =>
              ser.id === updateService.id
                ? updateService
                : ser
            )
          : [...(state.data ?? []), updateService],
      };
    }),
  setIsLoadingAllServices: (isLoadingAllServices) =>
    set({isLoadingAllServices}),
  setUpdateAllServices: (updateService) =>
    set((state) => {
      if (!updateService) return {data: state.data};
      const exists = state.data?.some(
        (ser) => ser?.id === updateService?.id
      );
      if (!exists) return {data: [...(state.data ?? [])]};
      return {
        data: state.data?.map((ser) =>
          ser.id === updateService.id ? updateService : ser
        ),
      };
    }),
  deleteServiceInData: (deleteServiceId) =>
    set((state) => {
      const exists = state.data?.some(
        (ser) => ser.id === deleteServiceId
      );

      return {
        data: exists
          ? state.data?.filter((ser) => ser.id !== deleteServiceId)
          : state.data,
      };
    }),
  addServiceInData: (service) =>
    set((state) => {
      return {data: [...(state.data ?? []), service]};
    }),
}));
