import { create } from "zustand";
import {
  type GetDashboardStatsResult,
  getDashboard,
} from "@/shared/api/dashboard/dashboard";

interface DashboardStore {
  dashboardInfo: GetDashboardStatsResult | null;
  setStats: (dashboardInfo: GetDashboardStatsResult | null) => void;

  loadStats: () => Promise<void>;
}

const initialState = {
  dashboardInfo: null,
};

export const useDashboardStore = create<DashboardStore>((set) => ({
  ...initialState,
  setStats: (dashboardInfo) => set({ dashboardInfo }),
  loadStats: async () => {
    const response = await getDashboard().getDashboardStats();
    set({ dashboardInfo: response });
  },
}));
