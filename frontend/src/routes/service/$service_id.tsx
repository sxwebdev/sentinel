import { createFileRoute } from "@tanstack/react-router";
import ServiceDetail from "@/pages/service/serviceDetail";

const ServiceDetailComponent = () => {
  const { service_id } = Route.useParams();
  return <ServiceDetail serviceID={service_id} />;
};

export const Route = createFileRoute("/service/$service_id")({
  component: ServiceDetailComponent,
});
