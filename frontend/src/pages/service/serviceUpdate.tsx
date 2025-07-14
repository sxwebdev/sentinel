import {ServiceForm} from "@/features/service/serviceForm";
import { useServiceUpdate } from "./hooks/useServiceUpdate";

export const ServiceUpdate = () => {
  const {serviceData, onUpdateService} = useServiceUpdate();
  return (
    <ServiceForm
      initialValues={serviceData}
      onSubmit={onUpdateService}
    />
  );
};