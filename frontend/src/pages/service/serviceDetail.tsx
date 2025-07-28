import ContentWrapper from "@/widgets/wrappers/contentWrapper";
import { useServiceDetail } from "./hooks/useServiceDetail";
import { Loader } from "@/entities/loader/loader";
import { ConfirmDialog } from "@/entities/confirmDialog/confirmDialog";
import { ServiceOverview } from "./components/serviceOverview";
import { ServiceStats } from "./components/serviceStats";
import { IncidentsList } from "./components/incidentsList";

const ServiceDetail = () => {
  const {
    filters,
    incidentsData,
    incidentsCount,
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
  } = useServiceDetail();

  if (!serviceDetailData || !incidentsData || !serviceStatsData)
    return <Loader loaderPage />;

  return (
    <ContentWrapper>
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
          incidentsCount={incidentsCount}
          filters={filters}
          setFilters={setFilters}
          setDeleteIncident={setDeleteIncident}
        />
      </div>
    </ContentWrapper>
  );
};

export default ServiceDetail;
