import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, Copy, Key, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { api } from "../api/client";

export function TokensPage() {
  const queryClient = useQueryClient();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [tokenName, setTokenName] = useState("");
  const [newlyCreatedToken, setNewlyCreatedToken] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["tokens"],
    queryFn: api.listTokens,
  });

  const tokens = data?.tokens ?? [];

  const createMutation = useMutation({
    mutationFn: api.createToken,
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ["tokens"] });
      setNewlyCreatedToken(res.token);
      setShowCreateModal(false);
      setTokenName("");
      setError(null);
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteToken,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tokens"] });
    },
  });

  const handleCopy = () => {
    if (newlyCreatedToken) {
      navigator.clipboard.writeText(newlyCreatedToken);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-white flex items-center gap-2">
            <Key className="h-6 w-6 text-indigo-400" /> API Tokens
          </h1>
          <p className="mt-1 text-sm text-slate-400">
            Manage personal access tokens for the Custodian CLI and CI/CD pipelines.
          </p>
        </div>
        <button
          onClick={() => {
            setError(null);
            setShowCreateModal(true);
          }}
          className="btn-primary flex items-center gap-2"
        >
          <Plus className="h-4 w-4" /> Issue token
        </button>
      </div>

      {isLoading && <p className="text-sm text-slate-400">Loading tokens...</p>}
      {isError && (
        <p className="text-sm text-rose-400">Failed to load API tokens.</p>
      )}

      {/* Newly Created Token Security Modal */}
      {newlyCreatedToken && (
        <div className="mb-6 rounded-lg border border-emerald-500/30 bg-emerald-950/20 p-5 backdrop-blur-sm">
          <div className="flex items-start justify-between">
            <div>
              <h3 className="font-semibold text-emerald-400">
                API Token Created Successfully!
              </h3>
              <p className="mt-1 text-xs text-slate-300">
                Please copy your token now. You will not be able to see it again!
              </p>
            </div>
            <button
              onClick={() => setNewlyCreatedToken(null)}
              className="text-xs text-slate-400 hover:text-white"
            >
              Dismiss
            </button>
          </div>
          <div className="mt-3 flex items-center gap-2">
            <code className="flex-1 rounded border border-emerald-800/50 bg-ink-950 px-3 py-2 font-mono text-sm text-emerald-300 select-all">
              {newlyCreatedToken}
            </code>
            <button
              onClick={handleCopy}
              className="flex items-center gap-1 rounded border border-emerald-500/40 bg-emerald-600/20 px-3 py-2 text-xs font-medium text-emerald-300 hover:bg-emerald-600/30 transition"
            >
              {copied ? (
                <>
                  <Check className="h-3.5 w-3.5" /> Copied!
                </>
              ) : (
                <>
                  <Copy className="h-3.5 w-3.5" /> Copy
                </>
              )}
            </button>
          </div>
        </div>
      )}

      {/* Tokens Table */}
      <div className="card overflow-hidden">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-ink-800 bg-ink-900 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Prefix</th>
              <th className="px-4 py-3">Created</th>
              <th className="px-4 py-3">Last Used</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {tokens.map((t) => (
              <tr
                key={t.id}
                className="border-b border-ink-800/60 last:border-0 hover:bg-ink-800/40"
              >
                <td className="px-4 py-3 font-medium text-white">{t.name}</td>
                <td className="px-4 py-3 font-mono text-xs text-slate-400">
                  {t.prefix || "cst_***"}
                </td>
                <td className="px-4 py-3 text-xs text-slate-500">
                  {new Date(t.created_at).toLocaleString()}
                </td>
                <td className="px-4 py-3 text-xs text-slate-500">
                  {t.last_used_at
                    ? new Date(t.last_used_at).toLocaleString()
                    : "Never"}
                </td>
                <td className="px-4 py-3 text-right">
                  <button
                    onClick={() => {
                      if (
                        confirm(`Are you sure you want to revoke token "${t.name}"?`)
                      ) {
                        deleteMutation.mutate(t.id);
                      }
                    }}
                    disabled={deleteMutation.isPending}
                    className="inline-flex items-center gap-1 text-xs text-rose-400 hover:text-rose-300"
                    title="Revoke Token"
                  >
                    <Trash2 className="h-3.5 w-3.5" /> Revoke
                  </button>
                </td>
              </tr>
            ))}
            {!isLoading && tokens.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-slate-500">
                  No API tokens issued yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Modal for Creating Token */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="w-full max-w-md card p-6 shadow-xl">
            <h2 className="text-xl font-semibold text-white">Issue API Token</h2>
            <p className="mt-1 text-xs text-slate-400">
              Enter a descriptive name for this token (e.g. "GitHub Actions CI" or "laptop-cli").
            </p>
            <form
              className="mt-4 space-y-4"
              onSubmit={(e) => {
                e.preventDefault();
                if (tokenName.trim()) {
                  createMutation.mutate(tokenName.trim());
                }
              }}
            >
              <div>
                <label className="label">Token Name</label>
                <input
                  className="input"
                  placeholder="cli-desktop"
                  value={tokenName}
                  onChange={(e) => setTokenName(e.target.value)}
                  autoFocus
                  required
                />
              </div>
              {error && <p className="text-xs text-rose-400">{error}</p>}
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-primary"
                  disabled={createMutation.isPending}
                >
                  {createMutation.isPending ? "Generating..." : "Generate Token"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
