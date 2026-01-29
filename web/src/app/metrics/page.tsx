"use client";

import Shell from "@/components/ui/Shell";
import { api } from "@/lib/api";
import { useInterval } from "@/lib/hooks";
import type { MetricsResponse, MetricsSeries } from "@/types";
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  TimeScale,
  Title,
  Tooltip,
} from "chart.js";
import { useState } from "react";
import { Line } from "react-chartjs-2";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  TimeScale,
);

const TIME_RANGES = [
  { value: "1h", label: "1 Hour" },
  { value: "24h", label: "24 Hours" },
  { value: "7d", label: "7 Days" },
  { value: "30d", label: "30 Days" },
];

const REFRESH_INTERVAL = 15_000;

export default function MetricsPage() {
  const [timeRange, setTimeRange] = useState("24h");
  const [metrics, setMetrics] = useState<MetricsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useInterval(async () => {
    try {
      const m = await api.getMetrics(timeRange);
      setMetrics(m);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load metrics");
    }
  }, REFRESH_INTERVAL, [timeRange]);

  return (
    <Shell>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Metrics</h1>
          <p className="mt-1 text-sm text-gray-500">
            Platform performance metrics &middot; refreshes every 15s
          </p>
        </div>
        <div className="flex gap-1 rounded-lg border border-gray-200 bg-white p-1">
          {TIME_RANGES.map((r) => (
            <button
              key={r.value}
              onClick={() => setTimeRange(r.value)}
              className={`rounded-md px-3 py-1.5 text-xs font-medium transition ${
                timeRange === r.value
                  ? "bg-kapok-600 text-white"
                  : "text-gray-600 hover:bg-gray-100"
              }`}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      {error && (
        <div className="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {metrics && (
        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
          <MetricChart
            title="Query Latency"
            series={[
              { ...metrics.query_latency_p50, label: "p50" },
              { ...metrics.query_latency_p95, label: "p95" },
              { ...metrics.query_latency_p99, label: "p99" },
            ]}
            yLabel="ms"
            colors={["#22c55e", "#f59e0b", "#ef4444"]}
          />
          <MetricChart
            title="Error Rate"
            series={[metrics.error_rate]}
            yLabel="%"
            colors={["#ef4444"]}
          />
          <MetricChart
            title="Throughput"
            series={[metrics.throughput]}
            yLabel="req/s"
            colors={["#3b82f6"]}
            className="lg:col-span-2"
          />
        </div>
      )}

      {!metrics && !error && (
        <div className="mt-20 text-center text-gray-400">Loading metrics...</div>
      )}
    </Shell>
  );
}

function MetricChart({
  title,
  series,
  yLabel,
  colors,
  className = "",
}: {
  title: string;
  series: MetricsSeries[];
  yLabel: string;
  colors: string[];
  className?: string;
}) {
  const data = {
    labels: series[0]?.data.map((d) => {
      const date = new Date(d.timestamp);
      return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
      });
    }) || [],
    datasets: series.map((s, i) => ({
      label: s.label,
      data: s.data.map((d) => d.value),
      borderColor: colors[i],
      backgroundColor: `${colors[i]}20`,
      fill: true,
      tension: 0.3,
      pointRadius: 0,
      borderWidth: 2,
    })),
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: "top" as const,
        labels: { boxWidth: 8, usePointStyle: true, pointStyle: "circle" },
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        title: { display: true, text: yLabel },
        grid: { color: "#f3f4f6" },
      },
      x: {
        grid: { display: false },
        ticks: { maxTicksLimit: 8 },
      },
    },
    interaction: {
      intersect: false,
      mode: "index" as const,
    },
  };

  return (
    <div
      className={`rounded-xl border border-gray-200 bg-white p-4 ${className}`}
    >
      <h3 className="text-sm font-semibold text-gray-700">{title}</h3>
      <div className="mt-3 h-64">
        <Line data={data} options={options} />
      </div>
    </div>
  );
}
