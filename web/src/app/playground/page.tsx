"use client";

import Shell from "@/components/ui/Shell";
import { api } from "@/lib/api";
import { useAsync } from "@/lib/hooks";
import type { Tenant } from "@/types";
import { useSearchParams } from "next/navigation";
import { Suspense, useCallback, useMemo, useState } from "react";

function PlaygroundContent() {
  const searchParams = useSearchParams();
  const initialTenant = searchParams.get("tenant") || "";
  const [selectedTenantId, setSelectedTenantId] = useState(initialTenant);
  const { data: tenants } = useAsync(() => api.listTenants());

  const graphqlEndpoint = useMemo(
    () => (selectedTenantId ? api.getGraphQLEndpoint(selectedTenantId) : ""),
    [selectedTenantId],
  );

  const fetcher = useCallback(
    async (params: { query: string; variables?: Record<string, unknown> }) => {
      const token = api.getAuthToken();
      const res = await fetch(graphqlEndpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(params),
      });
      return res.json();
    },
    [graphqlEndpoint],
  );

  return (
    <Shell>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">GraphQL Playground</h1>
          <p className="mt-1 text-sm text-gray-500">
            Test queries against your tenant&apos;s GraphQL API
          </p>
        </div>
        <select
          value={selectedTenantId}
          onChange={(e) => setSelectedTenantId(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-kapok-500 focus:outline-none focus:ring-1 focus:ring-kapok-500"
        >
          <option value="">Select tenant...</option>
          {tenants?.map((t: Tenant) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
      </div>

      <div className="mt-6 overflow-hidden rounded-xl border border-gray-200 bg-white">
        {!selectedTenantId ? (
          <div className="flex h-[500px] items-center justify-center text-gray-400">
            Select a tenant to start exploring its GraphQL API
          </div>
        ) : (
          <GraphiQLWrapper endpoint={graphqlEndpoint} fetcher={fetcher} />
        )}
      </div>
    </Shell>
  );
}

export default function PlaygroundPage() {
  return (
    <Suspense>
      <PlaygroundContent />
    </Suspense>
  );
}

function GraphiQLWrapper({
  endpoint,
  fetcher,
}: {
  endpoint: string;
  fetcher: (params: {
    query: string;
    variables?: Record<string, unknown>;
  }) => Promise<unknown>;
}) {
  const [query, setQuery] = useState(
    `# Welcome to Kapok GraphQL Playground
# Try a query against your tenant's API

query {
  __schema {
    types {
      name
    }
  }
}
`,
  );
  const [variables, setVariables] = useState("");
  const [result, setResult] = useState("");
  const [loading, setLoading] = useState(false);
  const [history, setHistory] = useState<string[]>([]);
  const [showHistory, setShowHistory] = useState(false);

  async function executeQuery() {
    setLoading(true);
    try {
      let vars: Record<string, unknown> | undefined;
      if (variables.trim()) {
        try {
          vars = JSON.parse(variables);
        } catch {
          setResult(JSON.stringify({ error: "Invalid JSON in variables" }, null, 2));
          return;
        }
      }
      const res = await fetcher({ query, variables: vars });
      setResult(JSON.stringify(res, null, 2));
      setHistory((prev) => [query, ...prev.slice(0, 19)]);
    } catch (e) {
      setResult(
        JSON.stringify(
          { error: e instanceof Error ? e.message : "Request failed" },
          null,
          2,
        ),
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex h-[600px] flex-col">
      <div className="flex items-center justify-between border-b border-gray-200 px-4 py-2">
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-500">{endpoint}</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowHistory(!showHistory)}
            className="rounded px-2 py-1 text-xs text-gray-500 hover:bg-gray-100"
          >
            History ({history.length})
          </button>
          <button
            onClick={executeQuery}
            disabled={loading}
            className="rounded-lg bg-kapok-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-kapok-700 disabled:opacity-50"
          >
            {loading ? "Running..." : "Run Query"}
          </button>
        </div>
      </div>

      {showHistory && history.length > 0 && (
        <div className="border-b border-gray-200 bg-gray-50 px-4 py-2">
          <p className="text-xs font-medium text-gray-500 mb-1">Query History</p>
          <div className="space-y-1 max-h-32 overflow-y-auto">
            {history.map((h, i) => (
              <button
                key={`${i}-${h.slice(0, 20)}`}
                onClick={() => {
                  setQuery(h);
                  setShowHistory(false);
                }}
                className="block w-full truncate rounded px-2 py-1 text-left text-xs text-gray-600 hover:bg-gray-100"
              >
                {h.split("\n").find((l) => l.trim() && !l.trim().startsWith("#")) || h.slice(0, 60)}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="grid flex-1 grid-cols-2 divide-x divide-gray-200 overflow-hidden">
        <div className="flex flex-col">
          <textarea
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="flex-1 resize-none bg-gray-50 p-4 font-mono text-sm focus:outline-none"
            spellCheck={false}
          />
          <div className="border-t border-gray-200">
            <p className="px-4 py-1 text-xs text-gray-400">Variables (JSON)</p>
            <textarea
              value={variables}
              onChange={(e) => setVariables(e.target.value)}
              className="h-20 w-full resize-none bg-gray-50 px-4 pb-2 font-mono text-xs focus:outline-none"
              placeholder='{}'
              spellCheck={false}
            />
          </div>
        </div>
        <pre className="overflow-auto bg-white p-4 font-mono text-sm text-gray-700">
          {result || "// Results will appear here"}
        </pre>
      </div>
    </div>
  );
}
