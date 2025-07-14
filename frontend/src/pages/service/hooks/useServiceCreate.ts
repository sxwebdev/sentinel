export const useServiceCreate = () => {
  const initialValues = {
    name: "",
    protocol: "",
    interval: 30,
    timeout: 10,
    retries: 3,
    tags: "",
    enabled: true,
    http_condition: "",
    endpoints: [
      {
        name: "",
        url: "",
        method: "GET",
        expected_status: 200,
      },
    ],
  };
  return {
    initialValues,
  };
};
