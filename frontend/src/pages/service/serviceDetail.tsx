import { useServiceDetail } from "./hooks/useServiceDetail";
import { Loader } from "@/entities/loader/loader";
import { ConfirmDialog } from "@/entities/confirmDialog/confirmDialog";
import { ServiceOverview } from "./components/serviceOverview";
import { ServiceStats } from "./components/serviceStats";
import { IncidentsList } from "./components/incidentsList";

type ServiceDetailProps = {
  serviceID: string;
};

const ServiceDetail = ({ serviceID }: ServiceDetailProps) => {
  const {
    filters,
    incidentsData,
    deleteIncident,
    resolveIncident,
    serviceDetailData,
    serviceStatsData,
    setFilters,
    onCheckService,
    setDeleteIncident,
    onDeleteIncident,
    setResolveIncident,
    onResolveIncident,
  } = useServiceDetail(serviceID);

  if (!serviceDetailData || !incidentsData || !serviceStatsData)
    return <Loader loaderPage />;

  return (
    <>
      <ConfirmDialog
        open={resolveIncident}
        setOpen={() => setResolveIncident(false)}
        onSubmit={onResolveIncident}
        title="Resolve Incident"
        description="Are you sure you want to resolve this incident?"
        type="default"
      />
      <ConfirmDialog
        open={!!deleteIncident}
        setOpen={() => setDeleteIncident(null)}
        onSubmit={() => onDeleteIncident(deleteIncident?.id ?? "")}
        title="Delete Incident"
        description="Are you sure you want to delete this incident?"
        type="delete"
      />

      <div className="flex flex-col gap-4 lg:gap-6">
        <ServiceOverview
          serviceDetailData={serviceDetailData}
          onCheckService={onCheckService}
          setResolveIncident={setResolveIncident}
        />

        <ServiceStats
          serviceDetailData={serviceDetailData}
          serviceStatsData={serviceStatsData}
        />

        <IncidentsList
          incidentsData={incidentsData}
          incidentsCount={incidentsData.count ?? 0}
          filters={filters}
          setFilters={setFilters}
          setDeleteIncident={setDeleteIncident}
        />
      </div>
    </>
  );
};

export default ServiceDetail;
