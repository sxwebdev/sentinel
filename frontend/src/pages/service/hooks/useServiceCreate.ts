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
      http: {
        condition: "",
        timeout: 10000,
        endpoints: [
          {
            name: "",
            url: "",
            expected_status: 200,
            method: "GET",
            body: "",
            headers: "",
            json_path: "",
            username: "",
            password: "",
          },
        ],
      },
      tcp: {
        endpoint: "",
        expect_data: "",
        send_data: "",
      },
      grpc: {
        endpoint: "",
        check_type: "health",
        tls: true,
        service_name: "",
        insecure_tls: false,
      },
    },
  };

  const onCreateService = async (values: ServiceForm) => {
    return await $api
      .post("/services", values)
      .then(() => {
        toast.success("Service created successfully");
        setIsOpenModal(false);
      })
      .catch((err) => {
        toast.error(err.data.error);
      });
  };
  return {
    initialValues,
    onCreateService,
    isOpenModal,
    setIsOpenModal,
  };
};
