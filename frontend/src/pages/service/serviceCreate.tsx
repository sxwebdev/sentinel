import { PlusIcon } from "lucide-react";
import { useRef, useState } from "react";
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
  DialogClose,
} from "@/shared/components/ui";
import {
  ServiceForm,
  type ServiceFormRef,
} from "@/features/service/serviceForm";
import { useServiceCreate } from "./hooks/useServiceCreate";
import { useIsMobile } from "@/shared/hooks/useIsMobile";

const ServiceCreate = () => {
  const { initialValues, onCreateService, isOpenModal, setIsOpenModal } =
    useServiceCreate();

  const isMobile = useIsMobile();

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
    <Dialog open={isOpenModal} onOpenChange={setIsOpenModal}>
      <DialogTrigger asChild>
        <Button size="sm">
          <PlusIcon />
          {isMobile ? "Add" : "Add Service"}
        </Button>
      </DialogTrigger>
      <DialogContent className="flex flex-col gap-0 p-0 max-h-[85vh] sm:max-h-[min(840px,95vh)] sm:max-w-2xl [&>button:last-child]:top-3.5">
        {/* Header */}
        <DialogHeader className="contents space-y-0 text-left">
          <DialogTitle className="border-b px-6 py-4 text-base">
            Create Service
          </DialogTitle>
        </DialogHeader>

        {/* Description */}
        <DialogDescription asChild>
          <div className="flex-1 overflow-y-auto overscroll-contain">
            <ServiceForm
              ref={formRef}
              initialValues={initialValues}
              onSubmit={onCreateService}
              type="create"
              onFormStateChange={setFormState}
            />
          </div>
        </DialogDescription>

        {/* Footer */}
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
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default ServiceCreate;
