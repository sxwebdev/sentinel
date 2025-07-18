import $api from "@/shared/api/baseApi";
import { useServiceTableStore } from "../store/useServiceTableStore";

export const useServiceApi = () => {
  const { setData } = useServiceTableStore();

  const onCheckService = async (id: string) => {
    await $api.post(`/services/${id}/check`);
  };

  const getAllServices = async () => {
    const res = await $api.get("/services");
    if (res.data === null) {
      setData([]);
    } else {
      setData(res.data);
    }
  };

  return {
    onCheckService,
    getAllServices,
  };
};
