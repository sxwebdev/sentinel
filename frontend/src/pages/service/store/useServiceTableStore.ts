import {create} from "zustand";
import type {Service} from "../../../features/service/types/type";

interface ServiceTableStore {
  data: Service[] | null;
  deleteService: Service | null;
  updateServiceId: string | null;
  isOpenDropdownIdAction: string | null;
  filters: {
    search: string;
    page: number;
  };
  setData: (data: Service[] | null) => void;
  setSearch: (search: string) => void;
  setPage: (page: number) => void;
  setDeleteService: (deleteService: Service | null) => void;
  setUpdateServiceId: (updateServiceId: string | null) => void;
  setIsOpenDropdownIdAction: (isOpenDropdownIdAction: string | null) => void;
}

const initialState = {
  data: null,
  deleteService: null,
  updateServiceId: null,
  isOpenDropdownIdAction: null,
  filters: {
    search: "",
    page: 1,
  },
};

export const useServiceTableStore = create<ServiceTableStore>((set) => ({
  ...initialState,
  setData: (data) => set({ data }),
  setIsOpenDropdownIdAction: (isOpenDropdownIdAction) =>
    set({isOpenDropdownIdAction}),
  setDeleteService: (deleteService) => set({deleteService}),
  setUpdateServiceId: (updateServiceId) => set({updateServiceId}),
  setSearch: (search) => set({filters: {...initialState.filters, search}}),
  setPage: (page) => set({filters: {...initialState.filters, page}}),
}));
