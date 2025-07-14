import {
  Badge,
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/shared/components/ui";
import {cn} from "@/shared/lib/utils";
import {
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from "@tanstack/react-table";
import {useEffect} from "react";
import {useNavigate} from "react-router";
import type {Service} from "../../../features/service/types/type";
import $api from "@/shared/api/baseApi";
import {
  EllipsisVerticalIcon,
  PencilIcon,
  RefreshCcwIcon,
  TrashIcon,
} from "lucide-react";
import {toast} from "sonner";
import {useServiceTableStore} from "../store/useServiceTableStore";
import {useServiceApi} from "./useServiceApi";
import {ActivityIndicatorSVG} from "@/entities/ActivityIndicatorSVG/ActivityIndicatorSVG";

export const useServiceTable = () => {
  const {
    data,
    filters,
    deleteService,
    setData,
    setSearch,
    setPage,
    setDeleteService,
    setUpdateServiceId,
  } = useServiceTableStore();

  const navigate = useNavigate();
  const {onCheckService} = useServiceApi();

  const onDeleteService = async () => {
    await $api
      .delete(`/services/${deleteService?.service?.id}`)
      .then(() => {
        toast.success("Service deleted");
      })
      .catch((err) => {
        toast.error(err.response.data.message);
      })
      .finally(() => {
        setDeleteService(null);
        getAllServices();
      });
  };

  const columns: ColumnDef<Service>[] = [
    {
      header: "Enabled",
      accessorKey: "enabled",
      cell: ({row}) => {
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
      cell: ({row}) => {
        return (
          <Button
            onClick={() => navigate(`/service/${row.original.service.id}`)}
            variant="link"
            className="cursor-pointer font-bold"
          >
            {row.original.service.name}
          </Button>
        );
      },
    },
    {
      header: "Status",
      accessorKey: "status",
      cell: ({row}) => {
        return (
          <Badge
            className={cn(
              row.original.state.status === "up" &&
                "bg-[#dcfce7] text-[#166534]",
              row.original.state.status === "down" &&
                "bg-[#fee2e2] text-[#991B1B]"
            )}
          >
            {row.original.state.status}
          </Badge>
        );
      },
    },
    {
      header: "Tags",
      accessorKey: "tags",
      cell: ({row}) => {
        return row.original.service.tags.join(", ") || "-";
      },
    },
    {
      header: "Last Check",
      accessorKey: "last_check",
      cell: ({row}) => {
        return new Date(row.original?.state?.last_check).toLocaleString("ru", {
          year: "numeric",
          month: "numeric",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        });
      },
    },
    {
      header: "Incidents",
      accessorKey: "incidents",
      cell: ({row}) => {
        return (
          <>
            <Badge variant="outline">
              {row.original.state.consecutive_fails ?? 0}
            </Badge>
            {" / "}
            <Badge variant="outline">
              {row.original.service.total_incidents ?? 0}
            </Badge>
          </>
        );
      },
    },
    {
      header: "Actions",
      accessorKey: "actions",
      cell: ({row}) => {
        return (
          <div className="flex justify-center">
            <DropdownMenu>
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
                  onClick={() => setDeleteService(row.original)}
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
  ];

  const getAllServices = async () => {
    const res = await $api.get("/services");
    if (res.data === null) {
      setData([]);
    } else {
      setData(res.data);
    }
  };

  const table = useReactTable({
    data: data ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  useEffect(() => {
    getAllServices();
  }, [filters]);

  return {
    onDeleteService,
    table,
    filters,
    setSearch,
    setPage,
    data,
    setData,
    deleteService,
    setDeleteService,
  };
};
