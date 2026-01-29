"use client";

import Shell from "@/components/ui/Shell";
import StatCard from "@/components/ui/StatCard";
import StatusBadge from "@/components/ui/StatusBadge";
import { api } from "@/lib/api";
import { formatBytes, formatNumber, formatRelative } from "@/lib/format";
import { useInterval } from "@/lib/hooks";
import type { PlatformStats, Tenant } from "@/types";
import Link from "next/link";
import { useState } from "react";

const REFRESH_INTERVAL = 30_000;

export default function DashboardPage() {
  const [stats, setStats] = useState<PlatformStats | null>(null);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [error, setError] = useState<string | null>(null);

  useInterval(async () => {
    try {
      const [s, t] = await Promise.all([api.getStats(), api.listTenants()]);
      setStats(s);
      setTenants(t);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load data");
    }
  }, REFRESH_INTERVAL);

  return (
    <Shell>
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="mt-1 text-sm text-gray-500">
        Platform overview &middot; auto-refreshes every 30s
      </p>

      {error && (
        <div className="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {stats && (
        <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard
            label="Total Tenants"
            value={formatNumber(stats.total_tenants)}
          />
          <StatCard
            label="Active Tenants"
            value={formatNumber(stats.active_tenants)}
          />
          <StatCard
            label="Total Storage"
            value={formatBytes(stats.total_storage_bytes)}
          />
          <StatCard
            label="Queries Today"
            value={formatNumber(stats.total_queries_today)}
          />
        </div>
      )}

      <div className="mt-8">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Tenants</h2>
          <Link
            href="/tenants"
            className="text-sm font-medium text-kapok-600 hover:text-kapok-700"
          >
            View all
          </Link>
        </div>

        <div className="mt-3 overflow-hidden rounded-xl border border-gray-200 bg-white">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-gray-500">Name</th>
                <th className="px-4 py-3 text-left font-medium text-gray-500">Status</th>
                <th className="px-4 py-3 text-left font-medium text-gray-500">Storage</th>
                <th className="px-4 py-3 text-left font-medium text-gray-500">Last Activity</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {tenants.map((t) => (
                <tr key={t.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link
                      href={`/tenants/${t.id}`}
                      className="font-medium text-kapok-600 hover:underline"
                    >
                      {t.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={t.status} />
                  </td>
                  <td className="px-4 py-3 text-gray-500">
                    {formatBytes(t.storage_used_bytes)}
                  </td>
                  <td className="px-4 py-3 text-gray-500">
                    {formatRelative(t.last_activity)}
                  </td>
                </tr>
              ))}
              {tenants.length === 0 && !error && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-gray-400">
                    No tenants yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </Shell>
  );
}
