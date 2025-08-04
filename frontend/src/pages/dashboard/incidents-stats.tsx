import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/components/ui/card";
import type { ChartConfig } from "@/shared/components/ui/chart";
import { ChartContainer, ChartTooltip } from "@/shared/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/components/ui/select";
import { getIncidents } from "@/shared/api/incidents/incidents";
import { useEffect, useMemo } from "react";
import { toast } from "sonner";

export const description = "An interactive area chart";

const ChartType = {
  IncidentsCount: "incidentsCount",
  IncidentsAvgDuration: "incidentsAvgDuration",
} as const;

type ChartType = (typeof ChartType)[keyof typeof ChartType];

interface IncidentStatsItem {
  date: string;
  count: number;
  avg_duration: number; // in seconds
  avg_duration_human: string; // for human-readable format
}

export function ChartIncidentsStats() {
  const [chartType, setChartType] = React.useState<ChartType>(
    ChartType.IncidentsCount
  );
  const [timeRange, setTimeRange] = React.useState("30d");
  const [incidentsData, setIncidentsData] = React.useState<IncidentStatsItem[]>(
    []
  );

  const currentChartConfig: ChartConfig = useMemo(
    () => ({
      incidents: {
        label:
          chartType === ChartType.IncidentsCount
            ? "Incidents Count"
            : "Avg Duration",
        color:
          chartType === ChartType.IncidentsCount
            ? "var(--chart-1)"
            : "var(--chart-2)",
      },
    }),
    [chartType]
  );

  useEffect(() => {
    const getDaysFromTimeRange = (range: string) => {
      switch (range) {
        case "7d":
          return 7;
        case "30d":
          return 30;
        case "90d":
          return 90;
        default:
          return 90;
      }
    };

    const days = getDaysFromTimeRange(timeRange);
    const endTime = new Date();
    const startTime = new Date();
    startTime.setDate(endTime.getDate() - days);

    getIncidents()
      .getIncidentsStats({
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
      })
      .then((response) => {
        const formattedData = response.map((item) => ({
          date: item.date || "",
          count: item.count || 0,
          avg_duration: item.avg_duration || 0,
          avg_duration_human: item.avg_duration_human || "",
        }));
        setIncidentsData(formattedData);
      })
      .catch((error) => {
        toast.error(
          "Error fetching incidents stats:",
          error?.message || "Unknown error"
        );
      });
  }, [timeRange]);

  const filteredData = useMemo(() => {
    return incidentsData.map((item) => {
      const date = new Date(item.date);
      const dataKey =
        chartType === ChartType.IncidentsCount
          ? item.count
          : Math.round(item.avg_duration);

      return {
        date: item.date,
        incidents: dataKey,
        count: item.count,
        avg_duration_human: item.avg_duration_human,
        formattedDate: date.toLocaleDateString("en-US", {
          month: "short",
          day: "numeric",
        }),
      };
    });
  }, [incidentsData, chartType]);

  return (
    <Card className="pt-0">
      <CardHeader className="flex md:items-center gap-2 py-6 md:py-0 space-y-0 border-b flex-col md:flex-row">
        <div className="grid flex-1 gap-1">
          <CardTitle>Incidents stats</CardTitle>
          <CardDescription>
            Showing total incidents for the last{" "}
            {timeRange === "90d"
              ? "3 months"
              : timeRange === "30d"
                ? "30 days"
                : "7 days"}
            .
          </CardDescription>
        </div>
        <Select
          value={chartType}
          onValueChange={(value) => setChartType(value as ChartType)}
        >
          <SelectTrigger
            className="w-full md:w-[180px] rounded-lg md:ml-auto"
            aria-label="Select a chart type"
          >
            <SelectValue placeholder="Incidents count" />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            <SelectItem value={ChartType.IncidentsCount} className="rounded-lg">
              Incidents count
            </SelectItem>
            <SelectItem
              value={ChartType.IncidentsAvgDuration}
              className="rounded-lg"
            >
              Average duration
            </SelectItem>
          </SelectContent>
        </Select>

        <Select value={timeRange} onValueChange={setTimeRange}>
          <SelectTrigger
            className="w-full md:w-[160px] rounded-lg sm:ml-auto "
            aria-label="Select a value"
          >
            <SelectValue placeholder="Last 3 months" />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            <SelectItem value="90d" className="rounded-lg">
              Last 3 months
            </SelectItem>
            <SelectItem value="30d" className="rounded-lg">
              Last 30 days
            </SelectItem>
            <SelectItem value="7d" className="rounded-lg">
              Last 7 days
            </SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer
          config={currentChartConfig}
          className="aspect-auto h-[250px] w-full"
        >
          <AreaChart data={filteredData}>
            <defs>
              <linearGradient id="fillIncidents" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor={
                    chartType === ChartType.IncidentsCount
                      ? "var(--chart-1)"
                      : "var(--chart-2)"
                  }
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor={
                    chartType === ChartType.IncidentsCount
                      ? "var(--chart-1)"
                      : "var(--chart-2)"
                  }
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
              tickFormatter={(value) => {
                const date = new Date(value);
                return date.toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                });
              }}
            />
            <ChartTooltip
              cursor={false}
              content={({ active, payload, label }) => {
                if (active && payload && payload.length && label) {
                  const data = payload[0].payload;
                  return (
                    <div className="rounded-lg border bg-background p-2 shadow-sm">
                      <div className="grid grid-cols-2 gap-2">
                        <div className="flex flex-col">
                          <span className="text-[0.70rem] uppercase text-muted-foreground">
                            Date
                          </span>
                          <span className="font-bold text-muted-foreground">
                            {new Date(label).toLocaleDateString("en-US", {
                              month: "short",
                              day: "numeric",
                            })}
                          </span>
                        </div>
                        <div className="flex flex-col">
                          <span className="text-[0.70rem] uppercase text-muted-foreground">
                            {chartType === ChartType.IncidentsCount
                              ? "Count"
                              : "Duration"}
                          </span>
                          <span className="font-bold">
                            {chartType === ChartType.IncidentsAvgDuration
                              ? data.avg_duration_human || "0s"
                              : data.count}
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                }
                return null;
              }}
            />
            <Area
              dataKey="incidents"
              type="natural"
              fill="url(#fillIncidents)"
              stroke={
                chartType === ChartType.IncidentsCount
                  ? "var(--chart-1)"
                  : "var(--chart-2)"
              }
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
