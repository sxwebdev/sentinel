import { systemClient } from "@/api/api";
import { ConnectError } from "@connectrpc/connect";
import { toast } from "sonner";
import { create } from "zustand";
import { create as createProto } from "@bufbuild/protobuf";
import {
  CheckIsInitializedRequestSchema,
  InitializeRequestSchema,
} from "@/api/gen/sentinel/system/v1/service_pb";

type SystemStore = {
  isLoading: boolean;
  isSystemInitialized: boolean;
  initializeSystem: (email: string, password: string) => Promise<void>;
};

export const useSystemStore = create<SystemStore>()((set) => {
  const store = {
    isLoading: true,
    isSystemInitialized: false,
    checkIsInitialized: async () => {
      try {
        const data = await systemClient.checkIsInitialized(
          createProto(CheckIsInitializedRequestSchema),
        );
        set({ isSystemInitialized: data.isInitialized });
      } catch (error) {
        if (error instanceof ConnectError) {
          toast.error(error.rawMessage);
        } else {
          console.error("Failed to check system initialization:", error);
        }
      } finally {
        set({ isLoading: false });
      }
    },
    initializeSystem: async (email: string, password: string) => {
      try {
        await systemClient.initialize(
          createProto(InitializeRequestSchema, {
            email,
            password,
          }),
        );
        set({ isSystemInitialized: true });
        toast.success("System initialized successfully");
      } catch (error) {
        if (error instanceof ConnectError) {
          toast.error(error.rawMessage);
        } else {
          console.error("Failed to initialize system:", error);
        }
      }
    },
  };

  store.checkIsInitialized();

  return store;
});
