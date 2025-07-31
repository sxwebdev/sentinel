import { create } from "zustand";
import type { GetDashboardStatsResult } from "@/shared/api/dashboard/dashboard";
import type { WebServerInfoResponse } from "@/shared/types/model";

interface DashboardStore {
  dashboardInfo: GetDashboardStatsResult | null;
  apiInfo: WebServerInfoResponse | null;
  setDashboardInfo: (dashboardInfo: GetDashboardStatsResult | null) => void;
  setApiInfo: (apiInfo: WebServerInfoResponse | null) => void;
}

const initialState = {
  dashboardInfo: null,
  apiInfo: null,
};

export const useDashboardStore = create<DashboardStore>((set) => ({
  ...initialState,
  setDashboardInfo: (dashboardInfo) => set({ dashboardInfo }),
  setApiInfo: (apiInfo) => set({ apiInfo }),
}));
