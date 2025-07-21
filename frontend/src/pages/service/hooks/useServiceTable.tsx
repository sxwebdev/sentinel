import {
  Badge,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
  Button,
} from "@/shared/components/ui";
import { cn } from "@/shared/lib/utils";
import {
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from "@tanstack/react-table";
import { useEffect, useMemo } from "react";
import { Link } from "react-router";
import type { Service } from "@features/service/types/type";
import $api from "@/shared/api/baseApi";
import {
  EllipsisIcon,
  PencilIcon,
  RefreshCcwIcon,
  TrashIcon,
} from "lucide-react";
import { toast } from "sonner";
import { useServiceTableStore } from "../store/useServiceTableStore";
import { useServiceApi } from "./useServiceApi";
import { ActivityIndicatorSVG } from "@/entities/ActivityIndicatorSVG/ActivityIndicatorSVG";

export const useServiceTable = () => {
  const {
    data,
    filters,
    servicesCount,
    deleteServiceId,
    isOpenDropdownIdAction,
    allTags,
    countAllTags,
    isLoadingAllServices,
    setData,
    setPage,
    setFilters,
    setServicesCount,
    setIsOpenDropdownIdAction,
    setIsLoadingAllServices,
    setAllTags,
    setCountAllTags,
    setDeleteServiceId,
    setUpdateServiceId,
  } = useServiceTableStore();

  const { onCheckService } = useServiceApi();

  const getAllTags = async () => {
    const res = await $api.get("/tags");
    setAllTags(res.data);
  };

  const getCountAllTags = async () => {
    const res = await $api.get("/tags/count");
    setCountAllTags(res.data);
  };

  const getAllServices = async () => {
    const res = await $api.get("/services", {
      params: {
        name: filters.search,
        page: filters.page,
        page_size: filters.pageSize,
        tags: filters.tags,
        protocol: filters.protocol,
        status: filters.status,
      },
    });
    if (res.data === null) {
      setData([]);
    } else {
      setData(res.data.items);
      setServicesCount(res.data.count);
    }
  };

  const onDeleteService = async () => {
    await $api
      .delete(`/services/${deleteServiceId}`)
      .then(() => {
        toast.success("Service deleted");
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      })
      .finally(() => {
        setDeleteServiceId(null);
      });
  };

  const columns: ColumnDef<Service>[] = useMemo(
    () => [
      {
        header: "Enabled",
        accessorKey: "enabled",
        size: 60,
        cell: ({ row }) => {
          return (
            <div className="flex items-center justify-center">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <ActivityIndicatorSVG
                      active={row.original?.is_enabled}
                      size={24}
                    />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>
                      {row.original?.is_enabled
                        ? "Service enabled"
                        : "Service disabled"}
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          );
        },
      },
      {
        header: "Service ",
        accessorKey: "service",
        cell: ({ row }) => {
          return (
            <Link
              to={`/service/${row.original?.id}`}
              className="cursor-pointer font-bold text-sm hover:underline"
            >
              {row.original?.name}
            </Link>
          );
        },
      },
      {
        header: "Status",
        accessorKey: "status",
        cell: ({ row }) => {
          return (
            <Badge
              className={cn(
                "text-xs font-semibold",
                row.original?.status === "up" &&
                  "bg-emerald-100 text-emerald-600",
                row.original?.status === "down" && "bg-rose-100 text-rose-600",
                row.original?.status === "unknown" &&
                  "bg-yellow-100 text-yellow-600"
              )}
            >
              {row.original?.status?.toUpperCase()}
            </Badge>
          );
        },
      },
      {
        header: "Tags",
        accessorKey: "tags",
        cell: ({ row }) => {
          if (row.original?.tags?.length === 0) {
            return (
              <div className="flex items-center justify-center">
                <div className="h-[3px] w-4 bg-gray-300 rounded-full" />
              </div>
            );
          }
          return (
            <div className="flex items-center justify-left flex-wrap gap-2">
              {row.original?.tags?.map((tag) => (
                <Badge
                  key={tag}
                  variant="outline"
                  className="text-sm font-medium"
                >
                  {tag}
                </Badge>
              ))}
            </div>
          );
        },
      },
      {
        header: "Last Check",
        accessorKey: "last_check",
        cell: ({ row }) => {
          return row.original?.last_check ? (
            new Date(row.original.last_check).toLocaleString("ru", {
              year: "numeric",
              month: "numeric",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
            })
          ) : (
            <div className="flex items-center justify-center">
              <div className="h-[3px] w-4 bg-gray-300 rounded-full" />
            </div>
          );
        },
      },
      {
        header: "Incidents",
        accessorKey: "incidents",
        cell: ({ row }) => {
          return (
            <>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <Badge variant="outline" className="text-xs font-medium">
                      {row.original?.active_incidents ?? 0}
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>Active incidents</TooltipContent>
                </Tooltip>
              </TooltipProvider>
              {" / "}
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <Badge variant="outline" className="text-xs font-medium">
                      {row.original?.total_incidents ?? 0}
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>Total incidents</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </>
          );
        },
      },
      {
        header: () => <span className="sr-only">Actions</span>,
        accessorKey: "actions",
        cell: ({ row }) => {
          return (
            <div>
              <DropdownMenu
                open={isOpenDropdownIdAction === row.original?.id}
                onOpenChange={(open) =>
                  open
                    ? setIsOpenDropdownIdAction(row.original?.id)
                    : setIsOpenDropdownIdAction(null)
                }
              >
                <DropdownMenuTrigger asChild>
                  <div className="flex justify-end">
                    <Button
                      size="icon"
                      variant="ghost"
                      className="shadow-none"
                      aria-label="Edit item"
                    >
                      <EllipsisIcon size={16} aria-hidden="true" />
                    </Button>
                  </div>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem
                    onClick={() => onCheckService(row.original?.id)}
                  >
                    <RefreshCcwIcon className="w-4 h-4" />
                    <span>Check</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => setUpdateServiceId(row.original?.id)}
                  >
                    <PencilIcon /> <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-destructive focus:text-destructive"
                    onClick={() => setDeleteServiceId(row.original?.id)}
                  >
                    <TrashIcon className="group-hover:text-white" />
                    <span>Delete</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          );
        },
      },
    ],
    [isOpenDropdownIdAction]
  );

  useEffect(() => {
    if (data === null) {
      setIsLoadingAllServices(true);
    }
    getAllServices().finally(() => {
      setIsLoadingAllServices(false);
    });
  }, [filters]);

  useEffect(() => {
    getAllTags();
    getCountAllTags();
  }, []);

  const table = useReactTable({
    data: data ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return {
    data,
    table,
    filters,
    servicesCount,
    allTags,
    countAllTags,
    deleteServiceId,
    isLoadingAllServices,
    setPage,
    setData,
    setFilters,
    onDeleteService,
    setDeleteServiceId,
  };
};
