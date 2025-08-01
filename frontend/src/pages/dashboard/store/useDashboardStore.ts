import { create } from "zustand";
import type { GetDashboardStatsResult } from "@/shared/api/dashboard/dashboard";

interface DashboardStore {
  dashboardInfo: GetDashboardStatsResult | null;
  setDashboardInfo: (dashboardInfo: GetDashboardStatsResult | null) => void;
}

const initialState = {
  dashboardInfo: null,
};

export const useDashboardStore = create<DashboardStore>((set) => ({
  ...initialState,
  setDashboardInfo: (dashboardInfo) => set({ dashboardInfo }),
}));
