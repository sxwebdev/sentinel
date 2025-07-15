import {Search} from "@/entities/search/search";
import {
  Card,
  CardContent,
  CardTitle,
  CardHeader,
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
} from "@/shared/components/ui";
import {flexRender} from "@tanstack/react-table";
import {useServiceTable} from "./hooks/useServiceTable";
import {Loader} from "@/entities/loader/loader";
import {PaginationBar} from "@/entities/paginationBar/paginationBar";
import { ConfirmDialog } from "@/entities/confirmDialog/confirmDialog";
import { ServiceUpdate } from "./serviceUpdate";

interface ServiceTableProps {
  onRefreshDashboard?: () => void;
}

export const ServiceTable = ({onRefreshDashboard}: ServiceTableProps) => {
  const {
    table,
    filters,
    setSearch,
    setPage,
    data,
    deleteService,
    setDeleteService,
    onDeleteService,
  } = useServiceTable();

  if (!data) return <Loader loaderPage />;

  return (
    <>
      <ServiceUpdate
        onRefreshDashboard={onRefreshDashboard}
      />
      <ConfirmDialog
        open={!!deleteService}
        setOpen={() => setDeleteService(null)}
        onSubmit={() => {
          onDeleteService().then(() => {
            onRefreshDashboard?.();
          });
        }}
        title="Delete Service"
        description="Are you sure you want to delete this service?"
        type="delete"
      />
      <Card>
        <CardHeader>
          <CardTitle>
            <h2 className="text-2xl font-bold">Services Overview</h2>
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          <div className="flex justify-between items-center">
            <Search
              placeholder="Search"
              value={filters.search}
              onChange={setSearch}
              clear
            />
          </div>
          <div className="rounded-xl overflow-hidden border border-border">
            <Table>
              <TableHeader className="bg-gray-100 rounded-t-lg">
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => {
                      return (
                        <TableHead key={header.id} className="text-center">
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
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      data-state={row.getIsSelected() && "selected"}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id} className="text-center">
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
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
              </TableBody>
            </Table>
          </div>
        </CardContent>
        <PaginationBar
          page={filters.page}
          setPage={setPage}
          total={data.length}
          pageSize={10}
          siblingCount={2}
        />
      </Card>
    </>
  );
};
