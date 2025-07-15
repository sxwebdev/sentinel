import type {ServiceForm} from "@/features/service/types/type";
import $api from "@/shared/api/baseApi";
import { useState } from "react";
import {toast} from "sonner";

export const useServiceCreate = () => {
  const [isOpenModal, setIsOpenModal] = useState(false);
  
  const initialValues = {
    name: "",
    protocol: "",
    interval: undefined,
    timeout: undefined,
    retries: undefined,
    tags: [],
    is_enabled: false,
    config: {
      http: {
        condition: "",
        timeout: undefined,
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
        check_type: "",
        tls: false,
        service_name: "",
        insecure_tls: false,
      },
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
