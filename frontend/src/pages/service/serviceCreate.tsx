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
import { DialogDescription } from "@radix-ui/react-dialog";
import { useServiceApi } from "./hooks/useServiceApi";

interface ServiceCreateProps {
  onRefreshDashboard: () => void;
}

const ServiceCreate = ({onRefreshDashboard}: ServiceCreateProps) => {
  const {getAllServices} = useServiceApi();
  const isMobile = useIsMobile();
  const {initialValues, onCreateService, isOpenModal, setIsOpenModal} =
    useServiceCreate();
  return (
    <Dialog open={isOpenModal} onOpenChange={setIsOpenModal}>
      <DialogTrigger asChild>
        <Button
          size="sm"
          className={cn(isMobile && "w-full")}
          variant="outline"
        >
          <PlusIcon />
          Add Service
        </Button>
      </DialogTrigger>
      <DialogDescription />
      <DialogContent className="overflow-y-auto max-h-[90vh]  sm:max-w-[90%] lg:max-w-[80%]">
        <DialogTitle>Create Service</DialogTitle>
        <hr />
        <ServiceForm
          initialValues={initialValues}
          onSubmit={async (values) => {
            return await onCreateService(values).then(() => {
              onRefreshDashboard();
              getAllServices();
            });

          }}
          type="create"
        />
      </DialogContent>
    </Dialog>
  );
};

export default ServiceCreate;
