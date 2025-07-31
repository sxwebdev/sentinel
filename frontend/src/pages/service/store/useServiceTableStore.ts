import {create} from "zustand";
import type {GetTagsCountResult, GetTagsResult} from "@/shared/api/tags/tags";
import type {GetServicesResult} from "@/shared/api/services/services";
import type {WebServiceDTO} from "@/shared/types/model";

interface ServiceTableStore {
  data: GetServicesResult | null;
  deleteServiceId: string | null;
  updateServiceId: string | null;
  createFromService: WebServiceDTO | null;
  isOpenDropdownIdAction: string | null;
  isLoadingAllServices: boolean;
  allTags: GetTagsResult | null;
  countAllTags: GetTagsCountResult | null;
  filters: {
    search: string | undefined;
    page: number;
    pageSize: number;
    tags: string[] | undefined;
    protocol: string | undefined;
    status: string | undefined;
  };
  setData: (data: GetServicesResult | null) => void;
  setFilters: (value: Partial<ServiceTableStore["filters"]>) => void;
  setUpdateService: (updateService: WebServiceDTO) => void;
  setPage: (page: number) => void;
  setAllTags: (allTags: GetTagsResult) => void;
  setCountAllTags: (countAllTags: GetTagsCountResult) => void;
  setCreateFromService: (createFromService: WebServiceDTO | null) => void;
  setUpdateAllServices: (updateService: WebServiceDTO) => void;
  setIsLoadingAllServices: (isLoadingAllServices: boolean) => void;
  setDeleteServiceId: (deleteServiceId: string | null) => void;
  setUpdateServiceId: (updateServiceId: string | null) => void;
  setIsOpenDropdownIdAction: (isOpenDropdownIdAction: string | null) => void;
  deleteServiceInData: (deleteServiceId: string) => void;
  addServiceInData: (service: WebServiceDTO) => void;
}

const initialState = {
  data: null,
  deleteServiceId: null,
  createFromService: null,
  updateServiceId: null,
  allTags: null,
  countAllTags: null,
  servicesCount: null,
  isOpenDropdownIdAction: null,
  isLoadingAllServices: false,
  filters: {
    search: undefined,
    page: 1,
    pageSize: 25,
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
  setCreateFromService: (createFromService) => set({createFromService}),
  setFilters: (value) =>
    set((state) => ({filters: {...state.filters, ...value}})),
  setUpdateServiceId: (updateServiceId) => set({updateServiceId}),
  setPage: (page) => set({ filters: { ...initialState.filters, page } }),
  setUpdateService: (updateService: WebServiceDTO) =>
    set((state) => {
      if (!updateService) return {data: state.data};
      const exists = state.data?.items?.some(
        (ser) => ser?.id === updateService?.id
      );
      return {
        data: {
          count: state.data?.count,
          items: exists
            ? state.data?.items?.map((ser) =>
                ser.id === updateService.id ? updateService : ser
              )
            : [...(state.data?.items ?? []), updateService],
        },
      };
    }),
  setIsLoadingAllServices: (isLoadingAllServices) =>
    set({isLoadingAllServices}),
  setUpdateAllServices: (updateService: WebServiceDTO) =>
    set((state) => {
      if (!updateService) return {data: state.data};
      const exists = state.data?.items?.some(
        (ser) => ser?.id === updateService?.id
      );
      if (!exists)
        return {
          data: {
            count: state.data?.count,
            items: [...(state.data?.items ?? [])],
          },
        };
      return {
        data: {
          count: state.data?.count,
          items: state.data?.items?.map((ser) =>
            ser.id === updateService.id ? updateService : ser
          ),
        },
      };
    }),
  deleteServiceInData: (deleteServiceId) =>
    set((state) => {
      const exists = state.data?.items?.some(
        (ser) => ser.id === deleteServiceId
      );

      return {
        data: {
          count: state.data?.count,
          items: exists
            ? state.data?.items?.filter((ser) => ser.id !== deleteServiceId)
            : state.data?.items,
        },
      };
    }),
  addServiceInData: (service: WebServiceDTO) =>
    set((state) => {
      return {
        data: {
          count: state.data?.count ? state.data.count + 1 : 1,
          items: [...(state.data?.items ?? []), service],
        },
      };
    }),
}));
