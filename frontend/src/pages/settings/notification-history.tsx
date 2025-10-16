import { notificationsClient } from "@/api/api";
import { historyList } from "@/api/gen/sentinel/notifications/v1/notifications-NotificationService_connectquery";
import { HistorySubscribeRequestSchema } from "@/api/gen/sentinel/notifications/v1/notifications_pb";
import PaginationTable from "@/shared/components/paginationTable";
import { H4, P } from "@/shared/components/typography";
import { Badge } from "@/shared/components/ui";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/shared/components/ui/popover";
import { Separator } from "@/shared/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/components/ui/table";
import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { useQuery } from "@connectrpc/connect-query";
import { useRef, useState } from "react";
import { useSubscriptionRefetch } from "@/shared/hooks/useSubscriptionRefetch";
import { toast } from "sonner";

const NotificationHistoryPage = () => {
  const [filters, setFilters] = useState({ page: 1, pageSize: 10 });
  const q = useQuery(historyList, {
    common: { page: filters.page, pageSize: filters.pageSize },
  });

  // setup subscription to refetch on new events
  const subRef = useRef<{ abort: () => void } | null>(null);
  useSubscriptionRefetch({
    refetch: q.refetch,
    subscribe: (signal) =>
      notificationsClient.historySubscribe(
        create(HistorySubscribeRequestSchema),
        { signal },
      ),
    debounceTimeout: 300,
    onErrorFn: (error) => {
      toast.error(`subscription error: ${error}`);
    },
    ref: subRef,
  });

  return (
    <>
      <H4 className="flex justify-between">Notification history</H4>
      <P className="text-muted-foreground text-sm leading-relaxed">
        On this page you can view the history of notifications sent by the
        system.
        <br />
        Old notifications are automatically deleted after 30 days.
      </P>
      <Separator className="my-5" />
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead className="w-[180px]">Message</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Attempts</TableHead>
              <TableHead>Sent at</TableHead>
              <TableHead>Created at</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {q.data?.items.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  No notifications found.
                </TableCell>
              </TableRow>
            ) : (
              q.data?.items?.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <Popover>
                      <PopoverTrigger>
                        <span className="cursor-pointer truncate text-sm">
                          {item.message.slice(0, 20) +
                            (item.message.length > 20 ? "..." : "")}
                        </span>
                      </PopoverTrigger>
                      <PopoverContent>
                        <div className="text-sm break-words">
                          {item.message.split("\n").map((line, i, arr) => (
                            <span key={i}>
                              {line}
                              {i < arr.length - 1 ? <br /> : null}
                            </span>
                          ))}
                        </div>
                      </PopoverContent>
                    </Popover>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={item.status === "sent" ? "success" : "info"}
                    >
                      {item.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{item.attempts}</TableCell>
                  <TableCell>
                    {item.sentAt
                      ? timestampDate(item.sentAt).toLocaleString()
                      : null}
                  </TableCell>
                  <TableCell>
                    {item.createdAt
                      ? timestampDate(item.createdAt).toLocaleString()
                      : null}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
      <PaginationTable
        selectedRows={filters.pageSize}
        setSelectedRows={(value) =>
          setFilters((prev) => ({ ...prev, pageSize: value }))
        }
        selectedPage={filters.page}
        setSelectedPage={(value) =>
          setFilters((prev) => ({ ...prev, page: value }))
        }
        totalPages={Math.ceil((q.data?.count ?? 0) / filters.pageSize)}
        className="mt-4 px-0"
      />
    </>
  );
};

export default NotificationHistoryPage;
