import { useState } from "react";
import { ShieldCheck } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuthStore } from "../store/auth";

export function LoginPage() {
  const [token, setToken] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const setTokenStore = useAuthStore((s) => s.setToken);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      setTokenStore(token.trim());
      await api.me();
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "authentication failed");
      useAuthStore.getState().clear();
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-ink-950 px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 flex items-center justify-center gap-2">
          <ShieldCheck className="h-8 w-8 text-accent-500" />
          <span className="text-2xl font-semibold text-white">Custodian</span>
        </div>
        <form onSubmit={submit} className="card space-y-4 p-6">
          <div>
            <label className="label">API token</label>
            <input
              className="input font-mono"
              placeholder="cst_ci.xxxx"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              autoFocus
            />
            <p className="mt-2 text-xs text-slate-500">
              Create one with <code className="text-slate-400">custodian tokens create</code>,
              or use your browser session via the OIDC login.
            </p>
          </div>
          {error && <p className="text-sm text-rose-400">{error}</p>}
          <button type="submit" className="btn-primary w-full justify-center" disabled={busy}>
            {busy ? "Verifying…" : "Sign in"}
          </button>
          <a href="/api/v1/auth/login" className="btn-secondary w-full justify-center">
            Continue with OIDC
          </a>
        </form>
      </div>
    </div>
  );
}
