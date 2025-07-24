import { create } from "zustand";
import type { GetDashboardStatsResult } from "@/shared/api/dashboard/dashboard";

interface DashboardStore {
  dashboardInfo: GetDashboardStatsResult | null;
  setDashboardInfo: (dashboardInfo: GetDashboardStatsResult | null) => void;
}

export const useDashboardStore = create<DashboardStore>((set) => ({
  dashboardInfo: null,
  setDashboardInfo: (dashboardInfo) => set({ dashboardInfo }),
}));
