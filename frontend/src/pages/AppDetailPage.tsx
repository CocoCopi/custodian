import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Code, FileCode2, History, Play, Save, Terminal, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { LogViewer } from "../components/LogViewer";
import { StatusBadge } from "../components/StatusBadge";
import { useLogStream } from "../hooks/useLogStream";
import { useLogStore } from "../store/logs";

export function AppDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [streaming, setStreaming] = useState(true);
  const [activeTab, setActiveTab] = useState<"logs" | "blueprint" | "history">("logs");
  const [blueprintText, setBlueprintText] = useState("");
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const queryClient = useQueryClient();

  const { data: svc, isLoading } = useQuery({
    queryKey: ["service", id],
    queryFn: () => api.getService(id!),
    enabled: !!id,
  });

  useEffect(() => {
    if (svc?.blueprint) {
      setBlueprintText(svc.blueprint);
    }
  }, [svc?.blueprint]);

  const { data: deploys } = useQuery({
    queryKey: ["deployments", id],
    queryFn: () => api.listDeployments(id!),
    enabled: !!id,
  });

  const deploy = useMutation({
    mutationFn: () => api.triggerDeploy(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["deployments", id] });
      queryClient.invalidateQueries({ queryKey: ["service", id] });
      setActiveTab("logs");
    },
  });

  const updateServiceMutation = useMutation({
    mutationFn: (newBp: string) => api.updateService(id!, { blueprint: newBp }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["service", id] });
      setSaveSuccess(true);
      setSaveError(null);
      setTimeout(() => setSaveSuccess(false), 3000);
    },
    onError: (err: Error) => {
      setSaveError(err.message);
    },
  });

  const deleteServiceMutation = useMutation({
    mutationFn: () => api.deleteService(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services"] });
      navigate("/apps");
    },
  });

  const entries = useLogStore((s) => s.entries);
  const connected = useLogStore((s) => s.connected);
  useLogStream(id, streaming);

  if (isLoading || !svc) {
    return <p className="text-sm text-slate-400">Loading…</p>;
  }

  const handleSaveAndDeploy = async () => {
    try {
      await updateServiceMutation.mutateAsync(blueprintText);
      deploy.mutate();
    } catch (e) {
      // error handled by mutation onError
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold text-white">{svc.name}</h1>
            <StatusBadge status={svc.status} />
          </div>
          <p className="mt-1 text-sm text-slate-400">
            {svc.repo_url || "local image"} · {svc.build_type} · branch {svc.branch}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="btn-secondary flex items-center gap-1.5 text-xs"
            onClick={() => setStreaming((v) => !v)}
          >
            {streaming ? "Pause logs" : "Resume logs"}
          </button>
          <button
            className="btn-primary flex items-center gap-1.5"
            onClick={() => deploy.mutate()}
            disabled={deploy.isPending}
          >
            <Play className="h-4 w-4 fill-current" />
            {deploy.isPending ? "Deploying…" : "Deploy"}
          </button>
          <button
            className="btn-secondary text-rose-400 hover:text-rose-300 p-2"
            title="Delete Application"
            onClick={() => {
              if (confirm(`Are you sure you want to delete application "${svc.name}"?`)) {
                deleteServiceMutation.mutate();
              }
            }}
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-ink-800 flex gap-6 text-sm font-medium">
        <button
          onClick={() => setActiveTab("logs")}
          className={`flex items-center gap-2 pb-3 border-b-2 transition ${
            activeTab === "logs"
              ? "border-accent-400 text-white"
              : "border-transparent text-slate-400 hover:text-slate-200"
          }`}
        >
          <Terminal className="h-4 w-4" /> Live Logs
        </button>
        <button
          onClick={() => setActiveTab("blueprint")}
          className={`flex items-center gap-2 pb-3 border-b-2 transition ${
            activeTab === "blueprint"
              ? "border-accent-400 text-white"
              : "border-transparent text-slate-400 hover:text-slate-200"
          }`}
        >
          <FileCode2 className="h-4 w-4" /> Blueprint Editor
        </button>
        <button
          onClick={() => setActiveTab("history")}
          className={`flex items-center gap-2 pb-3 border-b-2 transition ${
            activeTab === "history"
              ? "border-accent-400 text-white"
              : "border-transparent text-slate-400 hover:text-slate-200"
          }`}
        >
          <History className="h-4 w-4" /> Deploy History ({(deploys?.deployments ?? []).length})
        </button>
      </div>

      {/* Tab 1: Live Logs */}
      {activeTab === "logs" && (
        <div>
          <LogViewer entries={entries} connected={connected} />
        </div>
      )}

      {/* Tab 2: Blueprint Editor */}
      {activeTab === "blueprint" && (
        <div className="card p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                <Code className="h-5 w-5 text-indigo-400" /> custodian.yaml Blueprint
              </h2>
              <p className="mt-0.5 text-xs text-slate-400">
                Declarative specification for services, autoscaling, ingress domains, and volumes.
              </p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => updateServiceMutation.mutate(blueprintText)}
                disabled={updateServiceMutation.isPending}
                className="btn-secondary flex items-center gap-1.5 text-xs"
              >
                <Save className="h-3.5 w-3.5" /> Save Changes
              </button>
              <button
                onClick={handleSaveAndDeploy}
                disabled={updateServiceMutation.isPending || deploy.isPending}
                className="btn-primary flex items-center gap-1.5 text-xs"
              >
                <Play className="h-3.5 w-3.5 fill-current" /> Save & Deploy
              </button>
            </div>
          </div>

          {saveSuccess && (
            <p className="text-xs text-emerald-400 font-medium">✓ Blueprint saved successfully.</p>
          )}
          {saveError && (
            <p className="text-xs text-rose-400 font-medium">✗ {saveError}</p>
          )}

          <textarea
            className="w-full min-h-[420px] rounded-lg border border-ink-800 bg-ink-950 p-4 font-mono text-xs text-emerald-300 focus:border-indigo-500 focus:outline-none leading-relaxed"
            value={blueprintText}
            onChange={(e) => setBlueprintText(e.target.value)}
            placeholder="apiVersion: custodian.dev/v1..."
          />
        </div>
      )}

      {/* Tab 3: Deployments History */}
      {activeTab === "history" && (
        <div className="card overflow-hidden">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-ink-800 bg-ink-900 text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-4 py-3">Deployment ID</th>
                <th className="px-4 py-3">Commit SHA</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Created</th>
                <th className="px-4 py-3">Finished</th>
              </tr>
            </thead>
            <tbody>
              {(deploys?.deployments ?? []).map((d) => (
                <tr key={d.id} className="border-b border-ink-800/60 last:border-0 hover:bg-ink-800/40">
                  <td className="px-4 py-3 font-mono text-xs text-slate-300">{d.id.slice(0, 8)}</td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">
                    {d.commit_sha || "—"}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={d.status} />
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-400">
                    {new Date(d.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-500">
                    {d.finished_at ? new Date(d.finished_at).toLocaleString() : "Running / In Progress"}
                  </td>
                </tr>
              ))}
              {(deploys?.deployments ?? []).length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-slate-500">
                    No deployments recorded yet.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
