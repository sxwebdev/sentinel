import {create} from "zustand";
import type {DashboardInfo} from "../hooks/useDashboardLogic";

interface DashboardStore {
  dashboardInfo: DashboardInfo | null;
  setDashboardInfo: (dashboardInfo: DashboardInfo | null) => void;
}

export const useDashboardStore = create<DashboardStore>((set) => ({
  dashboardInfo: null,
  setDashboardInfo: (dashboardInfo) => set({dashboardInfo}),
}));