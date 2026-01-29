"use client";

import Shell from "@/components/ui/Shell";
import Modal from "@/components/ui/Modal";
import StatusBadge from "@/components/ui/StatusBadge";
import { api } from "@/lib/api";
import { formatBytes, formatDate, formatRelative } from "@/lib/format";
import { useAsync } from "@/lib/hooks";
import type { Tenant } from "@/types";
import Link from "next/link";
import { useState } from "react";

export default function TenantsPage() {
  const { data: tenants, loading, refetch } = useAsync(() => api.listTenants());
  const [showCreate, setShowCreate] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Tenant | null>(null);

  return (
    <Shell>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Tenants</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage your platform tenants
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-kapok-600 px-4 py-2 text-sm font-medium text-white hover:bg-kapok-700"
        >
          Create tenant
        </button>
      </div>

      <div className="mt-6 overflow-hidden rounded-xl border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Name</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Status</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Isolation</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Storage</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Created</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Last Activity</th>
              <th className="px-4 py-3 text-right font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {tenants?.map((t) => (
              <tr key={t.id} className="hover:bg-gray-50">
                <td className="px-4 py-3">
                  <Link
                    href={`/tenants/${t.id}`}
                    className="font-medium text-kapok-600 hover:underline"
                  >
                    {t.name}
                  </Link>
                </td>
                <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                <td className="px-4 py-3 text-gray-500">{t.isolation_level}</td>
                <td className="px-4 py-3 text-gray-500">{formatBytes(t.storage_used_bytes)}</td>
                <td className="px-4 py-3 text-gray-500">{formatDate(t.created_at)}</td>
                <td className="px-4 py-3 text-gray-500">{formatRelative(t.last_activity)}</td>
                <td className="px-4 py-3 text-right">
                  <button
                    onClick={() => setDeleteTarget(t)}
                    className="text-sm text-red-600 hover:text-red-700"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
            {!loading && (!tenants || tenants.length === 0) && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-gray-400">
                  No tenants yet. Create one to get started.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <CreateTenantModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
        }}
      />

      <DeleteTenantModal
        tenant={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onDeleted={() => {
          setDeleteTarget(null);
          refetch();
        }}
      />
    </Shell>
  );
}

function CreateTenantModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState("");
  const [isolation, setIsolation] = useState("schema");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await api.createTenant({ name, isolation_level: isolation });
      setName("");
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create tenant");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Create Tenant">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>
        )}
        <div>
          <label className="block text-sm font-medium text-gray-700">Name</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-kapok-500 focus:outline-none focus:ring-1 focus:ring-kapok-500"
            placeholder="my-tenant"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">
            Isolation Level
          </label>
          <select
            value={isolation}
            onChange={(e) => setIsolation(e.target.value)}
            className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-kapok-500 focus:outline-none focus:ring-1 focus:ring-kapok-500"
          >
            <option value="schema">Schema (recommended)</option>
            <option value="database">Database</option>
            <option value="row">Row-level</option>
          </select>
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="rounded-lg bg-kapok-600 px-4 py-2 text-sm font-medium text-white hover:bg-kapok-700 disabled:opacity-50"
          >
            {loading ? "Creating..." : "Create"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

function DeleteTenantModal({
  tenant,
  onClose,
  onDeleted,
}: {
  tenant: Tenant | null;
  onClose: () => void;
  onDeleted: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const [confirmName, setConfirmName] = useState("");
  const [error, setError] = useState("");

  async function handleDelete() {
    if (!tenant) return;
    setLoading(true);
    setError("");
    try {
      await api.deleteTenant(tenant.id);
      setConfirmName("");
      onDeleted();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete tenant");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal
      open={!!tenant}
      onClose={() => {
        setConfirmName("");
        setError("");
        onClose();
      }}
      title="Delete Tenant"
    >
      {error && (
        <p className="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>
      )}
      <p className="text-sm text-gray-600">
        This action cannot be undone. All data for tenant{" "}
        <strong>{tenant?.name}</strong> will be permanently deleted.
      </p>
      <p className="mt-3 text-sm text-gray-600">
        Type <strong>{tenant?.name}</strong> to confirm:
      </p>
      <input
        value={confirmName}
        onChange={(e) => setConfirmName(e.target.value)}
        className="mt-2 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-red-500 focus:outline-none focus:ring-1 focus:ring-red-500"
        placeholder={tenant?.name}
      />
      <div className="mt-4 flex justify-end gap-3">
        <button
          onClick={() => {
            setConfirmName("");
            onClose();
          }}
          className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          onClick={handleDelete}
          disabled={loading || confirmName !== tenant?.name}
          className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
        >
          {loading ? "Deleting..." : "Delete tenant"}
        </button>
      </div>
    </Modal>
  );
}
