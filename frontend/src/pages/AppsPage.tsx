import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { StatusBadge } from "../components/StatusBadge";

export function AppsPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["services"],
    queryFn: api.listServices,
  });
  const services = data?.services ?? [];

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-white">Applications</h1>
          <p className="mt-1 text-sm text-slate-400">
            Services deployed on your infrastructure.
          </p>
        </div>
        <Link to="/apps/new" className="btn-primary">
          New app
        </Link>
      </div>

      {isLoading && <p className="text-sm text-slate-400">Loading…</p>}
      {isError && (
        <p className="text-sm text-rose-400">Could not reach the control plane.</p>
      )}

      <div className="card overflow-hidden">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-ink-800 bg-ink-900 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Build</th>
              <th className="px-4 py-3">Branch</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Created</th>
            </tr>
          </thead>
          <tbody>
            {services.map((svc) => (
              <tr
                key={svc.id}
                className="border-b border-ink-800/60 last:border-0 hover:bg-ink-800/40"
              >
                <td className="px-4 py-3">
                  <Link to={`/apps/${svc.id}`} className="font-medium text-accent-400 hover:underline">
                    {svc.name}
                  </Link>
                </td>
                <td className="px-4 py-3 text-slate-400">{svc.build_type}</td>
                <td className="px-4 py-3 text-slate-400">{svc.branch}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={svc.status} />
                </td>
                <td className="px-4 py-3 text-xs text-slate-500">
                  {new Date(svc.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
            {!isLoading && services.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-slate-500">
                  No applications yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function NewAppPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [repo, setRepo] = useState("");
  const [buildType, setBuildType] = useState("dockerfile");
  const [blueprint, setBlueprint] = useState("");
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: api.createService,
    onSuccess: (svc) => {
      queryClient.invalidateQueries({ queryKey: ["services"] });
      navigate(`/apps/${svc.id}`);
    },
    onError: (e: Error) => setError(e.message),
  });

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-semibold text-white">New application</h1>
      <p className="mt-1 text-sm text-slate-400">
        Deploy from a git repository, an image, or a custodian.yaml blueprint.
      </p>

      <form
        className="card mt-6 space-y-5 p-6"
        onSubmit={(e) => {
          e.preventDefault();
          mutation.mutate({ name, repo_url: repo, build_type: buildType, blueprint });
        }}
      >
        <div>
          <label className="label">Name</label>
          <input
            className="input"
            placeholder="my-app"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div>
          <label className="label">Repository URL</label>
          <input
            className="input"
            placeholder="https://github.com/you/my-app.git"
            value={repo}
            onChange={(e) => setRepo(e.target.value)}
          />
        </div>
        <div>
          <label className="label">Build type</label>
          <select
            className="input"
            value={buildType}
            onChange={(e) => setBuildType(e.target.value)}
          >
            <option value="dockerfile">Dockerfile</option>
            <option value="buildpacks">Cloud Native Buildpacks</option>
            <option value="static">Static site</option>
          </select>
        </div>
        <div>
          <label className="label">Blueprint (custodian.yaml)</label>
          <textarea
            className="input min-h-40 font-mono text-xs"
            placeholder={"apiVersion: custodian.dev/v1\nkind: Blueprint\nname: my-app\nservices:\n  - name: web\n    build:\n      type: dockerfile"}
            value={blueprint}
            onChange={(e) => setBlueprint(e.target.value)}
          />
        </div>

        {error && <p className="text-sm text-rose-400">{error}</p>}
        <div className="flex items-center gap-3">
          <button type="submit" className="btn-primary" disabled={mutation.isPending}>
            {mutation.isPending ? "Creating…" : "Create app"}
          </button>
          <Link to="/apps" className="btn-secondary">
            Cancel
          </Link>
        </div>
      </form>
    </div>
  );
}
