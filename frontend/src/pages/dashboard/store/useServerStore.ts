import { create } from "zustand";
import type { WebServerInfoResponse } from "@/shared/types/model";
import { getServer } from "@/shared/api/server/server";
import { toast } from "sonner";

interface ServerStore {
  serverInfo: WebServerInfoResponse | null;
  isLoading: boolean;
  isUpdateAvailable: boolean;
  isUpdating: boolean;

  loadInfo: () => Promise<void>;
  doUpgrade: () => Promise<void>;
}

const initialState = {
  serverInfo: null,
  isLoading: false,
  isUpdateAvailable: false,
  isUpdating: false,
};

export const useServerStore = create<ServerStore>((set, get) => {
  const store = {
    ...initialState,
    loadInfo: async () => {
      try {
        set({ isLoading: true });
        const info = await getServer().getServerInfo();
        set({
          serverInfo: info,
          isLoading: false,
          isUpdateAvailable: info.available_update !== undefined,
        });
      } catch (error) {
        set({
          isLoading: false,
        });

        toast.error(error instanceof Error ? error.message : "Unknown error");
      }
    },
    doUpgrade: async () => {
      const { isUpdating, isUpdateAvailable } = get();

      if (isUpdating) {
        return;
      }

      if (!isUpdateAvailable) {
        toast.info("No updates available");
        return;
      }

      const toastID = toast.loading("Upgrading server...");

      try {
        set({ isUpdating: true });
        await getServer().getServerUpgrade();

        const interval = setInterval(() => {
          getServer()
            .getServerHealth()
            .then((health) => {
              if (health.status === "healthy") {
                toast.dismiss(toastID);

                set({ isUpdating: false, isUpdateAvailable: false });
                clearInterval(interval);
                toast.success("Server upgraded successfully!", {
                  description: "The page will refresh in 2 seconds.",
                  closeButton: true,
                  duration: Infinity,
                });
                setTimeout(() => {
                  window.location.reload();
                }, 2000);
              } else {
                toast.info("Server is still upgrading, please wait...");
              }
            })
            .catch(() => {
              // Ignore errors, will retry
            });
        }, 1000);
      } catch (error) {
        toast.dismiss(toastID);
        set({ isUpdating: false });
        toast.error("Failed to upgrade server", {
          description: error instanceof Error ? error.message : "Unknown error",
          closeButton: true,
          duration: Infinity,
        });
      }
    },
  };

  store.loadInfo();

  return store;
});
