import {ServiceForm} from "@/features/service/serviceForm";
import {
  Button,
  DialogTrigger,
  DialogContent,
  DialogDescription,
  DialogTitle,
  Dialog,
} from "@/shared/components/ui";
import {PlusIcon} from "lucide-react";
import {useCreateFromService} from "./hooks/useCreateFromService";

const ServiceCreateFromService = () => {
  const {
    createFromService,
    setCreateFromService,
    initialValues,
    onCreateService,
  } = useCreateFromService();

  if (!createFromService) return null;

  return (
    <Dialog
      open={!!createFromService}
      onOpenChange={() => setCreateFromService(null)}
    >
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

export default ServiceCreateFromService;
