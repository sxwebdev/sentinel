import type {ServiceForm} from "@/features/service/types/type";
import $api from "@/shared/api/baseApi";
import {useEffect, useState} from "react";
import {useShallow} from "zustand/react/shallow";
import {useServiceTableStore} from "../store/useServiceTableStore";
import {toast} from "sonner";

export const useServiceUpdate = () => {
  const [service, setService] = useState<ServiceForm | null>(null);
  const {updateServiceId, setUpdateServiceId} = useServiceTableStore(
    useShallow((s) => ({
      setUpdateServiceId: s.setUpdateServiceId,
      updateServiceId: s.updateServiceId,
    }))
  );

  const getService = async () => {
    await $api.get(`/services/${updateServiceId}`).then((res) => {
      setService(res.data);
    });
  };

  const onUpdateService = async (values: ServiceForm) => {
    await $api
      .put(`/services/${updateServiceId}`, values)
      .then(() => {
        toast.success("Service updated successfully");
        setUpdateServiceId(null);
      })
      .catch((err) => {
        toast.error(err.response.data.message);
      });
  };

  useEffect(() => {
    if (updateServiceId) {
      getService();
    }
  }, [updateServiceId]);

  return {service, setUpdateServiceId, onUpdateService};
};
