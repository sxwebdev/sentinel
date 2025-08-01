import { create } from "zustand";
import type { WebServerInfoResponse } from "@/shared/types/model";
import { getInfo } from "@/shared/api/info/info";

interface ServerStore {
  serverInfo: WebServerInfoResponse | null;
  isLoading: boolean;
  error: string | null;
  loadInfo: () => Promise<void>;
}

const initialState = {
  serverInfo: null,
  isLoading: false,
  error: null,
};

export const useServerStore = create<ServerStore>((set) => {
  const store = {
    ...initialState,
    loadInfo: async () => {
      try {
        set({ isLoading: true, error: null });
        const info = await getInfo().getInfo();
        set({ serverInfo: info, isLoading: false });
      } catch (error) {
        set({
          error: error instanceof Error ? error.message : "Unknown error",
          isLoading: false,
        });
      }
    },
  };

  store.loadInfo();

  return store;
});
