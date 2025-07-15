import type {ServiceForm} from "@/features/service/types/type";
import $api from "@/shared/api/baseApi";
import { useState } from "react";
import {toast} from "sonner";

export const useServiceCreate = () => {
  const [isOpenModal, setIsOpenModal] = useState(false);
  const initialValues: ServiceForm = {
    name: "",
    protocol: "",
    interval: 30000,
    timeout: 10000,
    retries: 3,
    tags: [],
    is_enabled: true,
    config: {
      grpc: null,
      tcp: null,
      http: null,
    },
  };

  const onCreateService = async (values: ServiceForm) => {
    await $api
      .post("/services", values)
      .then(() => {
        toast.success("Service created successfully");
        setIsOpenModal(false);
      })
      .catch((err) => {
        toast.error(err.message);
      });
  };
  return {
    initialValues,
    onCreateService,
    isOpenModal,
    setIsOpenModal,
  };
};
