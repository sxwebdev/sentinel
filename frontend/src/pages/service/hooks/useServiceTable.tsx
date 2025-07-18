import {
  Badge,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
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
  EllipsisVerticalIcon,
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
    deleteServiceId,
    isOpenDropdownIdAction,
    isLoadingAllServices,
    setData,
    setSearch,
    setPage,
    setIsOpenDropdownIdAction,
    setIsLoadingAllServices,
    setDeleteServiceId,
    setUpdateServiceId,
  } = useServiceTableStore();

  const { onCheckService } = useServiceApi();

    const getAllServices = async () => {
      const res = await $api.get("/services", {
        params: {
          filters: {
            search: filters.search,
          },
        },
      });
      if (res.data === null) {
        setData([]);
      } else {
        setData(res.data);
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
        cell: ({ row }) => {
          return (
            <div className="flex items-center justify-center">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <ActivityIndicatorSVG
                      active={row.original.service.is_enabled}
                      size={24}
                    />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>
                      {row.original.service.is_enabled ? "Enabled" : "Disabled"}
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
              to={`/service/${row.original.service.id}`}
              className="cursor-pointer font-bold text-sm hover:underline"
            >
              {row.original.service.name}
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
                "text-sm font-medium",
                row.original.state.status === "up" &&
                  "bg-green-light text-green",
                row.original.state.status === "down" && "bg-red-light text-red",
                row.original.state.status === "unknown" &&
                  "bg-orange-light text-orange",
              )}
            >
              {row.original.state.status.toUpperCase()}
            </Badge>
          );
        },
      },
      {
        header: "Tags",
        accessorKey: "tags",
        cell: ({ row }) => {
          if (row.original.service.tags.length === 0) {
            return (
              <div className="flex items-center justify-center">
                <div className="h-[3px] w-4 bg-gray-300 rounded-full" />
              </div>
            );
          }
          return (
            <div className="flex items-center justify-center flex-wrap gap-2">
              {row.original.service.tags.map((tag) => (
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
          return new Date(row.original?.state?.last_check).toLocaleString(
            "ru",
            {
              year: "numeric",
              month: "numeric",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
            },
          );
        },
      },
      {
        header: "Incidents",
        accessorKey: "incidents",
        cell: ({ row }) => {
          return (
            <>
              <Badge
                className={cn(
                  "text-sm font-medium",
                  row.original.service?.active_incidents > 0 &&
                    "bg-red-light text-red",
                  !row.original.service?.active_incidents &&
                    "bg-green-light text-green",
                )}
              >
                {row.original.service?.active_incidents ?? 0}
              </Badge>
              {" / "}
              <Badge variant="outline" className={cn("text-sm font-medium")}>
                {row.original.service.total_incidents ?? 0}
              </Badge>
            </>
          );
        },
      },
      {
        header: "Actions",
        accessorKey: "actions",
        cell: ({ row }) => {
          return (
            <div className="flex justify-center">
              <DropdownMenu
                open={isOpenDropdownIdAction === row.original.service.id}
                onOpenChange={(open) =>
                  open
                    ? setIsOpenDropdownIdAction(row.original.service.id)
                    : setIsOpenDropdownIdAction(null)
                }
              >
                <DropdownMenuTrigger className="cursor-pointer p-1">
                  <EllipsisVerticalIcon className="w-4 h-4" />
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem
                    onClick={() => onCheckService(row.original.service.id)}
                  >
                    <RefreshCcwIcon className="w-4 h-4" />
                    <span>Check</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => setUpdateServiceId(row.original.service.id)}
                  >
                    <PencilIcon /> <span>Edit</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="group focus:bg-destructive focus:text-white"
                    onClick={() => setDeleteServiceId(row.original.service.id)}
                  >
                    <TrashIcon className="text-muted-foreground group-hover:text-white" />
                    <span>Delete</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          );
        },
      },
    ],
    [isOpenDropdownIdAction],
  );

  useEffect(() => {
    setIsLoadingAllServices(true);
    getAllServices().finally(() => {
      setIsLoadingAllServices(false);
    });
  }, [filters.search]);

  const table = useReactTable({
    data: data ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return {
    isLoadingAllServices,
    onDeleteService,
    table,
    filters,
    setSearch,
    setPage,
    data,
    setData,
    deleteServiceId,
    setDeleteServiceId,
  };
};
