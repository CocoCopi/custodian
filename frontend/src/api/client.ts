import type { Deployment, Service } from "./types";

const API_BASE = import.meta.env.VITE_API_URL ?? "";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

function token(): string | null {
  return localStorage.getItem("custodian_token");
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string> | undefined),
  };
  const t = token();
  if (t) headers.Authorization = `Bearer ${t}`;

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      // non-JSON error body
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  me: () => request<{ owner_id: string }>("/api/v1/me"),

  listServices: () => request<{ services: Service[] }>("/api/v1/services"),
  createService: (body: {
    name: string;
    repo_url?: string;
    branch?: string;
    build_type?: string;
    blueprint?: string;
  }) =>
    request<Service>("/api/v1/services", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  getService: (id: string) => request<Service>(`/api/v1/services/${id}`),
  deleteService: (id: string) =>
    request<void>(`/api/v1/services/${id}`, { method: "DELETE" }),

  listDeployments: (serviceId: string) =>
    request<{ deployments: Deployment[] }>(`/api/v1/services/${serviceId}/deployments`),
  triggerDeploy: (serviceId: string, commit?: string) =>
    request<Deployment>(
      `/api/v1/services/${serviceId}/deployments${commit ? `?commit=${commit}` : ""}`,
      { method: "POST" },
    ),
};
