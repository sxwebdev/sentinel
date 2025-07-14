import $api from "@/shared/api/baseApi";

export const useServiceApi = () => {
    
  const onCheckService = async (id: string) => {
    await $api.post(`/services/${id}/check`);
  };

  return {
    onCheckService,
  };
};