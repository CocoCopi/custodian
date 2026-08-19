import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  CheckCircle2,
  Database,
  ExternalLink,
  HardDrive,
  Layers,
  RefreshCw,
  Server,
  ShieldCheck,
} from "lucide-react";

export function SystemHealthPage() {
  const { data: health, isLoading, isError, refetch } = useQuery({
    queryKey: ["healthz"],
    queryFn: async () => {
      const res = await fetch("/healthz");
      if (!res.ok) throw new Error("Health check failed");
      return res.json() as Promise<{ status: string; engine?: string; time?: string }>;
    },
    refetchInterval: 10000,
  });

  const getDomain = (sub: string) => {
    const host = window.location.hostname;
    const protocol = window.location.protocol;
    if (host === "localhost" || host === "127.0.0.1") {
      if (sub === "grafana") return `${protocol}//${host}:3000`;
      if (sub === "minio") return `${protocol}//${host}:9001`;
      if (sub === "traefik") return `${protocol}//${host}:8080/dashboard/`;
      return "#";
    }
    return `${protocol}//${sub}.${host}`;
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-white flex items-center gap-2">
            <Activity className="h-6 w-6 text-emerald-400" /> System & Infrastructure Health
          </h1>
          <p className="mt-1 text-sm text-slate-400">
            Real-time status of your self-hosted Custodian control plane and infrastructure components.
          </p>
        </div>
        <button
          onClick={() => refetch()}
          className="btn-secondary flex items-center gap-2 text-xs"
        >
          <RefreshCw className="h-3.5 w-3.5" /> Refresh
        </button>
      </div>

      {/* Main Status Overview */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase tracking-wider text-slate-400">
              Control Plane API
            </span>
            {isLoading ? (
              <span className="text-xs text-slate-400">Checking...</span>
            ) : isError ? (
              <span className="flex items-center gap-1 text-xs font-semibold text-rose-400">
                Offline
              </span>
            ) : (
              <span className="flex items-center gap-1 text-xs font-semibold text-emerald-400">
                <CheckCircle2 className="h-4 w-4" /> Healthy
              </span>
            )}
          </div>
          <div className="mt-4 flex items-center gap-3">
            <div className="rounded-lg bg-emerald-500/10 p-3 text-emerald-400">
              <Server className="h-6 w-6" />
            </div>
            <div>
              <p className="text-lg font-semibold text-white">Go Control API</p>
              <p className="text-xs text-slate-400">
                {health?.status === "ok" ? "All services operational" : "Connection issue"}
              </p>
            </div>
          </div>
        </div>

        <div className="card p-5">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase tracking-wider text-slate-400">
              Deployment Engine
            </span>
            <span className="rounded bg-indigo-500/20 px-2 py-0.5 text-xs font-semibold text-indigo-300 uppercase">
              {health?.engine ?? "compose"}
            </span>
          </div>
          <div className="mt-4 flex items-center gap-3">
            <div className="rounded-lg bg-indigo-500/10 p-3 text-indigo-400">
              <Layers className="h-6 w-6" />
            </div>
            <div>
              <p className="text-lg font-semibold text-white">
                {health?.engine === "k3s" ? "k3s Kubernetes" : "Docker Compose"}
              </p>
              <p className="text-xs text-slate-400">
                {health?.engine === "k3s"
                  ? "Multi-node HPA & KEDA support"
                  : "Single VPS Traefik routing"}
              </p>
            </div>
          </div>
        </div>

        <div className="card p-5">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase tracking-wider text-slate-400">
              Security & Auth
            </span>
            <span className="flex items-center gap-1 text-xs font-semibold text-emerald-400">
              <ShieldCheck className="h-4 w-4" /> Active
            </span>
          </div>
          <div className="mt-4 flex items-center gap-3">
            <div className="rounded-lg bg-purple-500/10 p-3 text-purple-400">
              <ShieldCheck className="h-6 w-6" />
            </div>
            <div>
              <p className="text-lg font-semibold text-white">OIDC & SHA-256 Tokens</p>
              <p className="text-xs text-slate-400">Role-based token isolation</p>
            </div>
          </div>
        </div>
      </div>

      {/* Services & Components Grid */}
      <h2 className="text-lg font-semibold text-white">Core Platform Subsystems</h2>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="card p-5 space-y-3">
          <div className="flex items-center gap-3">
            <Database className="h-5 w-5 text-indigo-400" />
            <div>
              <p className="text-sm font-semibold text-white">PostgreSQL Database</p>
              <p className="text-xs text-slate-400">
                Stores applications, deployment logs, and hashed API credentials.
              </p>
            </div>
          </div>
        </div>

        <div className="card p-5 space-y-3">
          <div className="flex items-center gap-3">
            <HardDrive className="h-5 w-5 text-emerald-400" />
            <div>
              <p className="text-sm font-semibold text-white">Redis & Asynq Job Queue</p>
              <p className="text-xs text-slate-400">
                Durable task queue managing asynchronous Docker & Buildpack tasks.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Observability Tools */}
      <h2 className="text-lg font-semibold text-white">Observability & Management Consoles</h2>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <a
          href={getDomain("grafana")}
          target="_blank"
          rel="noreferrer"
          className="card p-5 hover:border-indigo-500/50 transition group"
        >
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-white group-hover:text-indigo-400 flex items-center gap-2">
              Grafana Metrics <ExternalLink className="h-3.5 w-3.5" />
            </h3>
          </div>
          <p className="mt-2 text-xs text-slate-400">
            Prometheus metrics dashboard, CPU/memory graphs, and system usage.
          </p>
        </a>

        <a
          href={getDomain("minio")}
          target="_blank"
          rel="noreferrer"
          className="card p-5 hover:border-indigo-500/50 transition group"
        >
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-white group-hover:text-indigo-400 flex items-center gap-2">
              MinIO Storage <ExternalLink className="h-3.5 w-3.5" />
            </h3>
          </div>
          <p className="mt-2 text-xs text-slate-400">
            S3-compatible object storage console for static builds and assets.
          </p>
        </a>

        <a
          href={getDomain("traefik")}
          target="_blank"
          rel="noreferrer"
          className="card p-5 hover:border-indigo-500/50 transition group"
        >
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-white group-hover:text-indigo-400 flex items-center gap-2">
              Traefik Router <ExternalLink className="h-3.5 w-3.5" />
            </h3>
          </div>
          <p className="mt-2 text-xs text-slate-400">
            Edge reverse proxy routing dashboard and Let's Encrypt TLS certificates.
          </p>
        </a>
      </div>
    </div>
  );
}
