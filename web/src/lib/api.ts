import type {
  AuthTokens,
  MetricsResponse,
  PlatformStats,
  Tenant,
  User,
} from "@/types";

const API_URL = process.env.NEXT_PUBLIC_KAPOK_API_URL || "http://localhost:8080";

class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("kapok_access_token");
}

function setTokens(tokens: AuthTokens) {
  localStorage.setItem("kapok_access_token", tokens.access_token);
  localStorage.setItem("kapok_refresh_token", tokens.refresh_token);
}

function clearTokens() {
  localStorage.removeItem("kapok_access_token");
  localStorage.removeItem("kapok_refresh_token");
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });

  if (res.status === 401) {
    clearTokens();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
    throw new ApiError(401, "Unauthorized");
  }

  if (!res.ok) {
    const body = await res.text();
    throw new ApiError(res.status, body);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  // Auth
  async login(email: string, password: string): Promise<AuthTokens> {
    const tokens = await request<AuthTokens>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    setTokens(tokens);
    return tokens;
  },

  logout() {
    clearTokens();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  },

  async me(): Promise<User> {
    return request<User>("/api/v1/auth/me");
  },

  isAuthenticated(): boolean {
    return !!getToken();
  },

  // Platform stats
  async getStats(): Promise<PlatformStats> {
    return request<PlatformStats>("/api/v1/admin/stats");
  },

  // Tenants
  async listTenants(): Promise<Tenant[]> {
    return request<Tenant[]>("/api/v1/admin/tenants");
  },

  async getTenant(id: string): Promise<Tenant> {
    return request<Tenant>(`/api/v1/admin/tenants/${id}`);
  },

  async createTenant(data: {
    name: string;
    isolation_level: string;
  }): Promise<Tenant> {
    return request<Tenant>("/api/v1/admin/tenants", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async deleteTenant(id: string): Promise<void> {
    return request<void>(`/api/v1/admin/tenants/${id}`, {
      method: "DELETE",
    });
  },

  // Metrics
  async getMetrics(timeRange: string): Promise<MetricsResponse> {
    return request<MetricsResponse>(
      `/api/v1/admin/metrics?range=${timeRange}`,
    );
  },

  // GraphQL
  getGraphQLEndpoint(tenantId: string): string {
    return `${API_URL}/api/v1/tenants/${tenantId}/graphql`;
  },

  getAuthToken(): string | null {
    return getToken();
  },
};
