import { useState } from "react";
import { toast } from "sonner";
import type { WebCreateUpdateServiceRequest } from "@/shared/types/model";
import { getServices } from "@/shared/api/services/services";

export const useServiceCreate = () => {
  const [isOpenModal, setIsOpenModal] = useState(false);
  const { postServices } = getServices();

  // Create initial values
  const initialValues: WebCreateUpdateServiceRequest = {
    name: "",
    protocol: "http",
    interval: 60000,
    timeout: 10000,
    retries: 5,
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
            headers: {},
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

  // Handle create service
  const onCreateService = async (values: WebCreateUpdateServiceRequest) => {
    return await postServices(values)
      .then(() => {
        toast.success("Service created successfully");
        setIsOpenModal(false);
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
  };

  return {
    initialValues,
    onCreateService,
    isOpenModal,
    setIsOpenModal,
  };
};
