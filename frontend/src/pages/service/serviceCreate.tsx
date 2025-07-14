import {ServiceForm} from "@/features/service/serviceForm";
import {
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from "@/shared/components/ui";
import {useIsMobile} from "@/shared/hooks/useIsMobile";
import {cn} from "@/shared/lib/utils";
import {PlusIcon} from "lucide-react";
import {useServiceCreate} from "./hooks/useServiceCreate";

const ServiceCreate = () => {
  const isMobile = useIsMobile();
  const {initialValues} = useServiceCreate();
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button size="sm" className={cn(isMobile && "w-full")} variant="outline">
          <PlusIcon />
          Add Service
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90vh]">
        <DialogTitle>Create Service</DialogTitle>
        <hr />
        <ServiceForm initialValues={initialValues} />
      </DialogContent>
    </Dialog>
  );
};

export default ServiceCreate;
