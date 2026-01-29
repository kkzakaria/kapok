export interface Tenant {
  id: string;
  name: string;
  slug: string;
  isolation_level: string;
  status: "active" | "suspended" | "deleted";
  created_at: string;
  updated_at: string;
  storage_used_bytes: number;
  last_activity: string;
}

export interface PlatformStats {
  total_tenants: number;
  active_tenants: number;
  total_storage_bytes: number;
  total_queries_today: number;
}

export interface MetricsDataPoint {
  timestamp: string;
  value: number;
}

export interface MetricsSeries {
  label: string;
  data: MetricsDataPoint[];
}

export interface MetricsResponse {
  query_latency_p50: MetricsSeries;
  query_latency_p95: MetricsSeries;
  query_latency_p99: MetricsSeries;
  error_rate: MetricsSeries;
  throughput: MetricsSeries;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

export interface User {
  id: string;
  email: string;
  role: string;
}
