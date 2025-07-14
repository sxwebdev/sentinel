import {create} from "zustand";
import type {Service} from "../types/type";

interface ServiceTableStore {
  data: Service[] | null;
  deleteService: Service | null;
  filters: {
    search: string;
    page: number;
  };
  setData: (data: Service[] | null) => void;
  setSearch: (search: string) => void;
  setPage: (page: number) => void;
  setDeleteService: (deleteService: Service | null) => void;
}

const initialState = {
  data: null,
  deleteService: null,
  filters: {
    search: "",
    page: 1,
  },
};

export const useServiceTableStore = create<ServiceTableStore>((set) => ({
  ...initialState,
  setData: (data) => set({data}),
  setDeleteService: (deleteService) => set({deleteService}),
  setSearch: (search) => set({filters: {...initialState.filters, search}}),
  setPage: (page) => set({filters: {...initialState.filters, page}}),
}));
