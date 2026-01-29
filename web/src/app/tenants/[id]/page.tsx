"use client";

import Shell from "@/components/ui/Shell";
import StatusBadge from "@/components/ui/StatusBadge";
import { api } from "@/lib/api";
import { formatBytes, formatDate } from "@/lib/format";
import { useAsync } from "@/lib/hooks";
import Link from "next/link";
import { useParams } from "next/navigation";

export default function TenantDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: tenant, loading } = useAsync(() => api.getTenant(id), [id]);

  return (
    <Shell>
      <Link
        href="/tenants"
        className="text-sm text-gray-500 hover:text-gray-700"
      >
        &larr; Back to tenants
      </Link>

      {loading && <p className="mt-4 text-gray-400">Loading...</p>}

      {tenant && (
        <div className="mt-4">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold">{tenant.name}</h1>
            <StatusBadge status={tenant.status} />
          </div>

          <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2">
            <InfoRow label="ID" value={tenant.id} />
            <InfoRow label="Slug" value={tenant.slug} />
            <InfoRow label="Isolation" value={tenant.isolation_level} />
            <InfoRow label="Storage" value={formatBytes(tenant.storage_used_bytes)} />
            <InfoRow label="Created" value={formatDate(tenant.created_at)} />
            <InfoRow label="Updated" value={formatDate(tenant.updated_at)} />
          </div>

          <div className="mt-8">
            <h2 className="text-lg font-semibold">GraphQL Endpoint</h2>
            <code className="mt-2 block rounded-lg bg-gray-100 px-4 py-3 text-sm">
              {api.getGraphQLEndpoint(tenant.id)}
            </code>
            <Link
              href={`/playground?tenant=${tenant.id}`}
              className="mt-3 inline-block text-sm font-medium text-kapok-600 hover:text-kapok-700"
            >
              Open in Playground &rarr;
            </Link>
          </div>
        </div>
      )}
    </Shell>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white px-4 py-3">
      <p className="text-xs font-medium text-gray-500">{label}</p>
      <p className="mt-1 text-sm font-medium">{value}</p>
    </div>
  );
}
