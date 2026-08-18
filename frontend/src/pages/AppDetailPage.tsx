import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useParams } from "react-router-dom";
import { api } from "../api/client";
import { LogViewer } from "../components/LogViewer";
import { StatusBadge } from "../components/StatusBadge";
import { useLogStream } from "../hooks/useLogStream";
import { useLogStore } from "../store/logs";

export function AppDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [streaming, setStreaming] = useState(true);
  const queryClient = useQueryClient();

  const { data: svc, isLoading } = useQuery({
    queryKey: ["service", id],
    queryFn: () => api.getService(id!),
    enabled: !!id,
  });

  const { data: deploys } = useQuery({
    queryKey: ["deployments", id],
    queryFn: () => api.listDeployments(id!),
    enabled: !!id,
  });

  const deploy = useMutation({
    mutationFn: () => api.triggerDeploy(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["deployments", id] }),
  });

  const entries = useLogStore((s) => s.entries);
  const connected = useLogStore((s) => s.connected);
  useLogStream(id, streaming);

  if (isLoading || !svc) {
    return <p className="text-sm text-slate-400">Loading…</p>;
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold text-white">{svc.name}</h1>
            <StatusBadge status={svc.status} />
          </div>
          <p className="mt-1 text-sm text-slate-400">
            {svc.repo_url || "local image"} · {svc.build_type} · branch {svc.branch}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            className="btn-secondary"
            onClick={() => setStreaming((v) => !v)}
          >
            {streaming ? "Pause logs" : "Resume logs"}
          </button>
          <button className="btn-primary" onClick={() => deploy.mutate()} disabled={deploy.isPending}>
            {deploy.isPending ? "Deploying…" : "Deploy"}
          </button>
        </div>
      </div>

      <div className="mt-6">
        <LogViewer entries={entries} connected={connected} />
      </div>

      <div className="mt-6">
        <h2 className="mb-3 text-lg font-medium text-white">Deployments</h2>
        <div className="card overflow-hidden">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-ink-800 bg-ink-900 text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-4 py-3">ID</th>
                <th className="px-4 py-3">Commit</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Started</th>
              </tr>
            </thead>
            <tbody>
              {(deploys?.deployments ?? []).map((d) => (
                <tr key={d.id} className="border-b border-ink-800/60 last:border-0">
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">{d.id.slice(0, 8)}</td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">
                    {d.commit_sha || "—"}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={d.status} />
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-500">
                    {new Date(d.created_at).toLocaleString()}
                  </td>
                </tr>
              ))}
              {(deploys?.deployments ?? []).length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-slate-500">
                    No deployments yet — hit Deploy to ship your first one.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
