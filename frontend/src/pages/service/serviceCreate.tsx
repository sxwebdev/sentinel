import { ServiceForm } from "@/features/service/serviceForm";
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/shared/components/ui";
import { PlusIcon } from "lucide-react";
import { useServiceCreate } from "./hooks/useServiceCreate";

const ServiceCreate = () => {
  const { initialValues, onCreateService, isOpenModal, setIsOpenModal } =
    useServiceCreate();
  return (
    <Dialog open={isOpenModal} onOpenChange={setIsOpenModal}>
      <DialogTrigger asChild>
        <Button size="sm" variant="outline">
          <PlusIcon />
          Add Service
        </Button>
      </DialogTrigger>
      <DialogDescription />
      <DialogContent className="overflow-y-auto max-h-[90vh] sm:max-w-[90%] lg:max-w-[800px]">
        <DialogTitle>Create Service</DialogTitle>
        <hr />
        <ServiceForm
          initialValues={initialValues}
          onSubmit={onCreateService}
          type="create"
        />
      </DialogContent>
    </Dialog>
  );
};

export default ServiceCreate;
