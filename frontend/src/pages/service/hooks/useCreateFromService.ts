import { useMemo } from "react";
import { useServiceTableStore } from "../store/useServiceTableStore";
import { getServices } from "@/shared/api/services/services";
import { toast } from "sonner";
import type { WebCreateUpdateServiceRequest } from "@/shared/types/model";

export const useCreateFromService = () => {
  const { postServices } = getServices();
  const { createFromService, setCreateFromService } = useServiceTableStore();

  const initialValues = useMemo(() => {
    return {
      ...createFromService,
    };
  }, [createFromService]);

  // Handle create service
  const onCreateService = async (values: WebCreateUpdateServiceRequest) => {
    return await postServices(values)
      .then(() => {
        toast.success("Service created successfully");
        setCreateFromService(null);
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
  };

  return {
    initialValues,
    createFromService,
    setCreateFromService,
    onCreateService,
  };
};
