import { useQuery } from "@tanstack/react-query";
import { Boxes, Rocket, Activity } from "lucide-react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { StatusBadge } from "../components/StatusBadge";

export function Dashboard() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["services"],
    queryFn: api.listServices,
  });

  const services = data?.services ?? [];
  const running = services.filter((s) => s.status === "running").length;

  return (
    <div>
      <h1 className="text-2xl font-semibold text-white">Overview</h1>
      <p className="mt-1 text-sm text-slate-400">
        Self-hosted PaaS with full ownership of your infrastructure.
      </p>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard icon={Boxes} label="Applications" value={services.length} />
        <StatCard icon={Activity} label="Running" value={running} />
        <StatCard icon={Rocket} label="Deploys" value="—" />
      </div>

      <div className="mt-8">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-medium text-white">Applications</h2>
          <Link to="/apps/new" className="btn-primary">
            New app
          </Link>
        </div>

        {isLoading && <p className="text-sm text-slate-400">Loading…</p>}
        {isError && (
          <p className="text-sm text-rose-400">
            Could not reach the control plane. Check your API token.
          </p>
        )}
        {!isLoading && !isError && services.length === 0 && (
          <div className="card p-8 text-center">
            <p className="text-sm text-slate-400">
              No applications yet. Create your first app to get started.
            </p>
          </div>
        )}

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {services.map((svc) => (
            <Link
              key={svc.id}
              to={`/apps/${svc.id}`}
              className="card p-4 transition-colors hover:border-ink-700"
            >
              <div className="flex items-center justify-between">
                <span className="font-medium text-white">{svc.name}</span>
                <StatusBadge status={svc.status} />
              </div>
              <p className="mt-2 truncate text-xs text-slate-500">
                {svc.repo_url || svc.image || svc.build_type}
              </p>
              <p className="mt-1 text-xs text-slate-600">
                branch: {svc.branch} · {svc.build_type}
              </p>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Boxes;
  label: string;
  value: number | string;
}) {
  return (
    <div className="card p-5">
      <div className="flex items-center gap-3">
        <div className="rounded-lg bg-ink-800 p-2.5">
          <Icon className="h-5 w-5 text-accent-400" />
        </div>
        <div>
          <p className="text-xs uppercase tracking-wide text-slate-500">{label}</p>
          <p className="text-2xl font-semibold text-white">{value}</p>
        </div>
      </div>
    </div>
  );
}
