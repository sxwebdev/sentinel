import type {ServiceForm} from "@/features/service/types/type";
import $api from "@/shared/api/baseApi";
import {useEffect, useState} from "react";
import {useShallow} from "zustand/react/shallow";
import {useServiceTableStore} from "../store/useServiceTableStore";
import {toast} from "sonner";

export const useServiceUpdate = () => {
    const [serviceData, setServiceData] = useState<ServiceForm | null>(null);
    const [isLoading, setIsLoading] = useState(false);
  const {updateServiceId, setUpdateServiceId} = useServiceTableStore(
    useShallow((s) => ({
      setUpdateServiceId: s.setUpdateServiceId,
      updateServiceId: s.updateServiceId,
    }))
  );

  const getService = async () => {
    setIsLoading(true);
      await $api.get(`/services/${updateServiceId}`).then((res) => {
      setServiceData(res.data.service);
    }).finally(() => {
      setIsLoading(false);
    });
  };

  const onUpdateService = async (values: ServiceForm) => {
    return await $api
      .put(`/services/${updateServiceId}`, values)
      .then(() => {
        toast.success("Service updated successfully");
        setUpdateServiceId(null);
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
  };

  useEffect(() => {
    if (updateServiceId) {
      getService();
    }
  }, [updateServiceId]);

  return {
    serviceData,
    setUpdateServiceId,
    onUpdateService,
    isLoading,
    updateServiceId,
  };
};
