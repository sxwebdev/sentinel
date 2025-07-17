import {ServiceForm} from "@/features/service/serviceForm";
import {useServiceUpdate} from "./hooks/useServiceUpdate";
import {Loader} from "@/entities/loader/loader";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/shared/components/ui";
import {useServiceApi} from "./hooks/useServiceApi";
import { cn } from "@/shared/lib/utils";

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
  return (
    <Dialog
      open={!!updateServiceId}
      onOpenChange={() => {
        setUpdateServiceId(null);
      }}
    >
      <DialogDescription />

      <DialogContent
        className={cn(
          "overflow-y-auto max-h-[90%]  sm:max-w-[90%] lg:max-w-[800px] h-full",
          isLoading && "flex flex-col"
        )}
      >
        <DialogTitle className="h-fit">Update Service</DialogTitle>
        <hr className="h-fit" />
        {isLoading ? (
          <div className="flex items-center justify-center h-full w-full">
            <Loader />
          </div>
        ) : (
          <>
            {serviceData ? (
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
            ) : (
              <div className="flex items-center justify-center h-full">
                <p>Service not found</p>
              </div>
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};
