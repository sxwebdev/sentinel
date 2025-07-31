import {
  Card,
  CardContent,
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
  SelectItem,
  SelectWithClear,
} from "@/shared/components/ui";
import {flexRender} from "@tanstack/react-table";
import {useServiceTable} from "./hooks/useServiceTable";
import {Loader} from "@/entities/loader/loader";
import {ConfirmDialog} from "@/entities/confirmDialog/confirmDialog";
import {ServiceUpdate} from "./serviceUpdate";
import {cn} from "@/shared/lib/utils";
import {Search} from "@/entities/search/search";
import PaginationTable from "@/shared/components/paginationTable";
import MultiSelect from "@/shared/components/multiSelect";

interface ServiceTableProps {
  protocols: Record<string, number>;
}

export const ServiceTable = ({protocols}: ServiceTableProps) => {
  const {
    data,
    table,
    filters,
    allTags,
    countAllTags,
    setFilters,
    deleteServiceId,
    setDeleteServiceId,
    onDeleteService,
    isLoadingAllServices,
  } = useServiceTable();

  return (
    <>
      <ServiceUpdate />
      <ConfirmDialog
        open={!!deleteServiceId}
        setOpen={() => setDeleteServiceId(null)}
        onSubmit={onDeleteService}
        title="Delete Service"
        description="Are you sure you want to delete this service?"
        type="delete"
      />
      <Card>
        <CardContent className="flex flex-col gap-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 justify-between items-center w-full gap-3">
            <Search
              className="lg:col-span-2"
              placeholder="Search"
              value={filters.search}
              onChange={(value) => setFilters({search: value ?? undefined})}
              clear
            />
            <MultiSelect
              className="lg:col-span-2"
              options={
                allTags?.map((tag) => ({
                  label: `${tag} (${countAllTags?.[tag] ?? 0})`,
                  value: tag,
                })) ?? []
              }
              value={
                filters.tags?.map((tag) => ({
                  label: tag,
                  value: tag,
                })) ?? []
              }
              onChange={(value) => {
                setFilters({tags: value.map((v) => v.value)});
              }}
              placeholder="Select tags"
            />
            <SelectWithClear
              className="w-full"
              value={filters.protocol || ""}
              onValueChange={(value) => {
                setFilters({protocol: value || undefined});
              }}
              onClear={() => setFilters({protocol: undefined})}
              placeholder="Select protocol"
            >
              {Object.keys(protocols).map((protocol) => {
                return (
                  <SelectItem key={protocol} value={protocol}>
                    {protocol}
                  </SelectItem>
                );
              })}
              {Object.keys(protocols).length === 0 && (
                <SelectItem value="none" disabled>
                  No protocols available
                </SelectItem>
              )}
            </SelectWithClear>
            <SelectWithClear
              className="w-full"
              value={filters.status || ""}
              onValueChange={(value) => {
                setFilters({status: value || undefined});
              }}
              onClear={() => setFilters({status: undefined})}
              placeholder="Select status"
            >
              <SelectItem value="up">Up</SelectItem>
              <SelectItem value="down">Down</SelectItem>
            </SelectWithClear>
          </div>
          <div className="rounded-xl overflow-hidden border border-border">
            <Table>
              <TableHeader className="bg-gray-100 rounded-t-lg">
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header, idx) => {
                      return (
                        <TableHead
                          key={header.id}
                          className={cn(idx === 0 && "w-0 whitespace-nowrap")}
                        >
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext()
                              )}
                        </TableHead>
                      );
                    })}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {isLoadingAllServices ? (
                  <TableRow>
                    <TableCell
                      colSpan={table.getAllColumns().length}
                      className="h-24"
                    >
                      <Loader size={6} />
                    </TableCell>
                  </TableRow>
                ) : (
                  <>
                    {table.getRowModel().rows?.length ? (
                      table.getRowModel().rows.map((row) => {
                        return (
                          <TableRow
                            key={row.original?.id}
                            data-state={row.getIsSelected() && "selected"}
                          >
                            {row.getVisibleCells().map((cell) => (
                              <TableCell key={cell.id}>
                                {flexRender(
                                  cell.column.columnDef.cell,
                                  cell.getContext()
                                )}
                              </TableCell>
                            ))}
                          </TableRow>
                        );
                      })
                    ) : (
                      <TableRow>
                        <TableCell
                          colSpan={table.getAllColumns().length}
                          className="h-24 text-center"
                        >
                          No services found.
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
        <PaginationTable
          selectedRows={filters.pageSize}
          setSelectedRows={(value) => setFilters({pageSize: value})}
          selectedPage={filters.page}
          setSelectedPage={(value) => setFilters({page: value})}
          totalPages={Math.ceil((data?.count ?? 0) / filters.pageSize)}
        />
      </Card>
    </>
  );
};
