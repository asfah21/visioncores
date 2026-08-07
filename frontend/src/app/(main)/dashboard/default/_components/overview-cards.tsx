"use client";

import { useEffect, useState } from "react";

import { format, subMonths } from "date-fns";
import { TrendingDown, TrendingUp } from "lucide-react";
import { Bar, BarChart, Line, LineChart, XAxis } from "recharts";

import { Card, CardAction, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { cn } from "@/lib/utils";

import { leadsChartConfig } from "./crm.config";

const trafficChartData = [
  { time: "09 AM", total: 10, known: 5, unknown: 0 },
  { time: "10 AM", total: 12, known: 6, unknown: 2 },
  { time: "12 PM", total: 10, known: 10, unknown: 4 },
  { time: "13 PM", total: 6, known: 12, unknown: 3 },
  { time: "14 PM", total: 11, known: 11, unknown: 6 },
  { time: "16 PM", total: 14, known: 11, unknown: 4 },
];

const trafficChartConfig = {
  total: { label: "Total", color: "#6366f1" },
  known: { label: "People", color: "#22c55e" },
  unknown: { label: "Unknown", color: "#f97316" },
};

const _lastMonth = format(subMonths(new Date(), 1), "LLLL");

import { UserCheck, UserMinus, Users } from "lucide-react";

import { Badge } from "@/components/ui/badge";

interface Detection {
  name: string;
  person_id: number;
  bbox: { x1: number; y1: number; x2: number; y2: number };
  frame_size: { w: number; h: number };
  camera_id?: string;
}

interface DailyData {
  date: string;
  newLeads: number;
  disqualified: number;
}

interface HeatmapData {
  time: string;
  total: number;
  known: number;
  unknown: number;
}

const formatHour = (hour: number) => {
  const h = hour % 12 || 12;
  const ampm = hour < 12 ? "AM" : "PM";
  return `${String(h).padStart(2, "0")} ${ampm}`;
};

export function OverviewCards({ className, cameraId }: { className?: string; cameraId: string }) {
  const [liveDetections, setLiveDetections] = useState<Detection[]>([]);
  const [totalToday, setTotalToday] = useState(0);
  const [dailyData, setDailyData] = useState<DailyData[]>([]);
  const [heatmapData, setHeatmapData] = useState<HeatmapData[]>([]);
  const [peakHourInfo, setPeakHourInfo] = useState({ hour: "00 AM", count: 0 });
  const [growth, setGrowth] = useState(0);

  const API = "http://10.10.11.5:3001";

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [_c, d, h] = await Promise.all([
          fetch(`${API}/count?camera_id=${cameraId}`).then((r) => r.json()),
          fetch(`${API}/daily?camera_id=${cameraId}`).then((r) => r.json()),
          fetch(`${API}/heatmap?camera_id=${cameraId}`).then((r) => r.json()),
        ]);

        // Pastikan angka Total Today selaras dengan akumulasi di Heatmap
        const heatmapTotal = (h || []).reduce((sum: number, item: { count: number }) => sum + item.count, 0);
        setTotalToday(heatmapTotal);

        // Map API daily data to chart format
        const mappedDaily = (d || []).slice(-6).map((item: { date: string; count: number }) => ({
          date: item.date.split("T")[0],
          newLeads: item.count,
          disqualified: Math.floor(item.count * 0.1),
        }));
        setDailyData(mappedDaily);

        // Map API heatmap data to traffic chart format for today (progressively per hour)
        const currentHour = new Date().getHours();
        const mappedTraffic = Array.from({ length: Math.max(1, currentHour + 1) }).map((_, i) => {
          const apiItem = (h || []).find((x: { hour: number; count: number }) => x.hour === i);
          const count = apiItem ? apiItem.count : 0;
          return {
            time: formatHour(i),
            total: count,
            known: count, // Using full count to match UI naming "People"
            unknown: 0,
          };
        });

        // Hapus jam yang nilainya nol beruntun agar X-axis tidak terlalu padat
        const filteredTraffic = mappedTraffic.filter((item, i, arr) => {
          if (item.known === 0 && i > 0 && arr[i - 1].known === 0) {
            return false; // Skip jika sebelumnya sudah nol
          }
          return true;
        });

        setHeatmapData(filteredTraffic);

        // Find peak hour
        if (h && h.length > 0) {
          const peak = [...h].sort((a, b) => b.count - a.count)[0];
          setPeakHourInfo({
            hour: formatHour(peak.hour),
            count: peak.count,
          });
        }

        // Calculate growth if we have at least 2 days
        if (d && d.length >= 2) {
          const today = d[d.length - 1].count;
          const yesterday = d[d.length - 2].count;
          if (yesterday > 0) {
            const perc = ((today - yesterday) / yesterday) * 100;
            setGrowth(Number(perc.toFixed(1)));
          } else {
            setGrowth(0);
          }
        } else {
          setGrowth(0);
        }
      } catch (err) {
        console.error("API ERROR:", err);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [cameraId]);

  useEffect(() => {
    const handleDetections = (e: Event) => {
      const detail = ((e as CustomEvent).detail || []) as Detection[];
      // Filter live detections for the current camera
      const filtered = detail.filter((d) => d.camera_id === cameraId || !d.camera_id);
      setLiveDetections(filtered);
    };
    window.addEventListener("ai:detections", handleDetections);
    return () => window.removeEventListener("ai:detections", handleDetections);
  }, [cameraId]);

  const peopleCount = liveDetections.length;
  const facesCount = liveDetections.filter((d) => d.name && d.name !== "Unknown").length;
  const unknownCount = liveDetections.filter((d) => d.name === "Unknown").length;
  return (
    <div
      className={cn(
        "grid grid-cols-1 gap-0 *:data-[slot=card]:shadow-none sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6",
        className,
      )}
    >
      <Card className="flex h-full flex-col @container/card">
        <CardHeader>
          <CardTitle>Daily Stats</CardTitle>
        </CardHeader>
        <CardContent>
          <ChartContainer className="h-18 w-full" config={leadsChartConfig}>
            <BarChart accessibilityLayer data={dailyData} barSize={8}>
              <XAxis dataKey="date" tickLine={false} tickMargin={10} axisLine={false} hide />
              <ChartTooltip content={<ChartTooltipContent labelFormatter={(label) => `Date: ${label}`} />} />
              <Bar
                background={{ fill: "var(--color-background)", radius: 2, opacity: 0.07 }}
                dataKey="newLeads"
                stackId="a"
                fill="var(--color-newLeads)"
                radius={[0, 0, 0, 0]}
              />
              {/* <Bar dataKey="disqualified" stackId="a" fill="var(--color-disqualified)" radius={[2, 2, 0, 0]} /> */}
            </BarChart>
          </ChartContainer>
        </CardContent>
        <CardFooter className="mt-auto flex items-center justify-between">
          <span className={cn("font-bold text-blue-300 text-lg")}> Today </span>
          <span className="font-semibold text-xl tabular-nums">
            {totalToday} <span className="font-normal text-sm">people</span>
          </span>
        </CardFooter>
      </Card>

      <Card className="flex h-full flex-col @container/card">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Heatmap</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <ChartContainer className="h-20 w-full" config={trafficChartConfig}>
            <LineChart
              data={heatmapData.length > 0 ? heatmapData : trafficChartData}
              margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
            >
              <XAxis
                dataKey="time"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                className="text-[10px] opacity-50"
              />
              <ChartTooltip content={<ChartTooltipContent />} />
              {/* <Line
                type="monotone"
                dataKey="total"
                stroke="var(--color-total)"
                strokeWidth={2}
                dot={{ r: 3, fill: "var(--color-total)" }}
              /> */}
              <Line
                type="monotone"
                dataKey="known"
                stroke="var(--color-known)"
                strokeWidth={2}
                dot={{ r: 3, fill: "var(--color-known)" }}
              />
              {/* <Line
                type="monotone"
                dataKey="unknown"
                stroke="var(--color-unknown)"
                strokeWidth={2}
                dot={{ r: 3, fill: "var(--color-unknown)" }}
              /> */}
            </LineChart>
          </ChartContainer>
        </CardContent>
        <CardFooter className="mt-auto flex justify-center gap-4 font-medium text-[12px] opacity-60">
          {/* <div className="flex items-center gap-1">
            <div className="size-1.5 rounded-full bg-[#6366f1]" /> Total
          </div> */}
          <div className="flex items-center gap-1">
            <div className="size-1.5 rounded-full bg-[#22c55e]" /> People
          </div>
          {/* <div className="flex items-center gap-1">
            <div className="size-1.5 rounded-full bg-[#f97316]" /> Unknown
          </div> */}
        </CardFooter>
      </Card>

      <Card className="flex h-full flex-col @container/card">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Peak Traffic</CardTitle>
          </div>
          {/* <CardDescription>Peak Traffic</CardDescription> */}
          <CardTitle className="pt-2 pb-1 font-semibold @[250px]/card:text-3xl text-2xl tabular-nums">
            <span className="font-bold text-green-500 text-4xl">{peakHourInfo.count}</span>
            <span className="font-normal text-sm"> people</span>
          </CardTitle>
          <CardAction>
            {/* <Badge variant="outline">
              {growth >= 0 ? <TrendingUp className="text-green-500" /> : <TrendingDown className="text-red-500" />}
              {growth >= 0 ? `+${growth}%` : `${growth}%`}
            </Badge> */}
            <Badge variant="outline" className={growth >= 0 ? "text-green-500" : "text-red-500"}>
              {growth >= 0 ? <TrendingUp className="text-green-500" /> : <TrendingDown className="text-red-500" />}
              {growth >= 0 ? `+${growth}%` : `${growth}%`}
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="mt-auto flex-col items-start gap-1.5 text-sm">
          <div className="line-clamp-1 flex gap-2 font-medium">Peak Hour: {peakHourInfo.hour}</div>
          <div className="text-muted-foreground">(Highest traffic)</div>
        </CardFooter>
      </Card>

      <Card className="flex h-full flex-col">
        <CardHeader>
          <CardTitle>Live Presence</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="group flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="rounded-md bg-green-500/15 p-1.5 text-green-600">
                <Users className="size-3.5" />
              </div>
              <span className="font-medium text-xs opacity-70">People detected</span>
            </div>
            <span className="font-bold text-sm tabular-nums">{peopleCount}</span>
          </div>

          <div className="group flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="rounded-md bg-blue-500/15 p-1.5 text-blue-600">
                <UserCheck className="size-3.5" />
              </div>
              <span className="font-medium text-xs opacity-70">Faces detected</span>
            </div>
            <span className="font-bold text-sm tabular-nums">{facesCount}</span>
          </div>

          <div className="group flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="rounded-md bg-destructive/15 p-1.5 text-destructive">
                <UserMinus className="size-3.5" />
              </div>
              <span className="font-medium text-xs opacity-70">Unknown faces</span>
            </div>
            <span className="font-bold text-destructive text-sm tabular-nums">{unknownCount}</span>
          </div>
        </CardContent>
      </Card>

      {/* <Card className="col-span-1 xl:col-span-2">
        <CardHeader>
          <CardTitle>Revenue Growth</CardTitle>
          <CardDescription>Year to Date (YTD)</CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer config={revenueChartConfig} className="h-24 w-full">
            <LineChart
              data={revenueChartData}
              margin={{
                top: 5,
                right: 10,
                left: 10,
                bottom: 0,
              }}
            >
              <XAxis dataKey="month" tickLine={false} tickMargin={10} axisLine={false} hide />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Line
                type="monotone"
                strokeWidth={2}
                dataKey="revenue"
                stroke="var(--color-revenue)"
                activeDot={{
                  r: 6,
                }}
              />
            </LineChart>
          </ChartContainer>
        </CardContent>
        <CardFooter>
          <p className="text-muted-foreground text-sm">+35% growth since last year</p>
        </CardFooter>
      </Card> */}
    </div>
  );
}
