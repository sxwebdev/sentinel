import {ServiceForm} from "@/features/service/serviceForm";
import {useServiceUpdate} from "./hooks/useServiceUpdate";
import {Loader} from "@/entities/loader/loader";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/shared/components/ui";
import { useServiceApi } from "./hooks/useServiceApi";

interface ServiceUpdateProps {
  onRefreshDashboard?: () => void;

}

export const ServiceUpdate = ({onRefreshDashboard}: ServiceUpdateProps) => {
  const {getAllServices} = useServiceApi();
  const {
    serviceData,
    onUpdateService,
    isLoading,
    updateServiceId,
    setUpdateServiceId,
  } = useServiceUpdate();
  if (isLoading) return <Loader />;

  return (
    <Dialog
      open={!!updateServiceId}
      onOpenChange={() => {
        setUpdateServiceId(null);
      }}
    >
      <DialogDescription />
      {serviceData ? (
        <DialogContent className="overflow-y-auto max-h-[90vh]  sm:max-w-[90%] lg:max-w-[80%]">
          <DialogTitle>Update Service</DialogTitle>
          <hr />
          <ServiceForm
            type="update"
            initialValues={serviceData}
            onSubmit={async (values) => {
              return await onUpdateService(values).then(() => {
                onRefreshDashboard?.();
                getAllServices?.();
              });
            }}
          />
        </DialogContent>
      ) : (
        <DialogContent>
          <DialogTitle>Service not found</DialogTitle>
          <DialogDescription>Service not found</DialogDescription>
        </DialogContent>
      )}
    </Dialog>
  );
};
