import { useRef, useState } from "react";
import {
  ServiceForm,
  type ServiceFormRef,
} from "@/features/service/serviceForm";
import { useServiceUpdate } from "./hooks/useServiceUpdate";
import { Loader } from "@/entities/loader/loader";
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogDescription,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from "@/shared/components/ui";

import { cn } from "@/shared/lib/utils";

export const ServiceUpdate = () => {
  const {
    serviceData,
    onUpdateService,
    isLoading,
    updateServiceId,
    setUpdateServiceId,
  } = useServiceUpdate();

  const formRef = useRef<ServiceFormRef>(null);

  const [formState, setFormState] = useState({
    isSubmitting: false,
    isValid: false,
    dirty: false,
  });

  const handleSubmit = () => {
    formRef.current?.submitForm();
  };

  return (
    <Dialog
      open={!!updateServiceId}
      onOpenChange={() => {
        setUpdateServiceId(null);
      }}
    >
      <DialogContent
        className={cn(
          "flex flex-col gap-0 p-0 max-h-[85vh] sm:max-h-[min(840px,95vh)] sm:max-w-2xl [&>button:last-child]:top-3.5",
          (isLoading || !serviceData) && "flex flex-col"
        )}
      >
        <DialogHeader className="contents space-y-0 text-left">
          <DialogTitle className="border-b px-6 py-4 text-base">
            Update Service
          </DialogTitle>
        </DialogHeader>
        <DialogDescription asChild>
          <div className="flex-1 overflow-y-auto overscroll-contain">
            {isLoading ? (
              <div className="flex items-center justify-center py-10 w-full">
                <Loader />
              </div>
            ) : (
              <>
                {serviceData ? (
                  <ServiceForm
                    ref={formRef}
                    initialValues={serviceData}
                    onSubmit={onUpdateService}
                    type="update"
                    onFormStateChange={setFormState}
                  />
                ) : (
                  <div className="flex items-center justify-center h-full w-full">
                    <p>Service not found</p>
                  </div>
                )}
              </>
            )}
          </div>
        </DialogDescription>
        <DialogFooter className="flex-shrink-0 border-t px-6 py-4 sm:items-center">
          <DialogClose asChild>
            <Button variant="outline" type="button">
              Cancel
            </Button>
          </DialogClose>
          <Button
            type="button"
            onClick={handleSubmit}
            disabled={
              formState.isSubmitting || !formState.isValid || !formState.dirty
            }
          >
            Update
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
