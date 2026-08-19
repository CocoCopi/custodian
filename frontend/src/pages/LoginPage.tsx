import { useEffect, useState } from "react";
import { KeyRound, Lock, Mail, ShieldCheck, User, UserPlus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuthStore } from "../store/auth";

export function LoginPage() {
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null);
  const [tab, setTab] = useState<"local" | "token">("local");

  // Login form state
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [token, setToken] = useState("");

  // Registration form state
  const [regUsername, setRegUsername] = useState("");
  const [regEmail, setRegEmail] = useState("");
  const [regPassword, setRegPassword] = useState("");
  const [regConfirmPassword, setRegConfirmPassword] = useState("");

  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const setTokenStore = useAuthStore((s) => s.setToken);
  const navigate = useNavigate();

  useEffect(() => {
    api
      .getSetupStatus()
      .then((res) => setSetupRequired(res.setup_required))
      .catch(() => setSetupRequired(false));
  }, []);

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (regPassword !== regConfirmPassword) {
      setError("Passwords do not match");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await api.registerUser(regUsername.trim(), regPassword, regEmail.trim());
      setTokenStore(res.token);
      await api.me();
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
      useAuthStore.getState().clear();
    } finally {
      setBusy(false);
    }
  };

  const handleLocalSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const res = await api.localLogin(username.trim(), password);
      setTokenStore(res.token);
      await api.me();
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid credentials");
      useAuthStore.getState().clear();
    } finally {
      setBusy(false);
    }
  };

  const handleTokenSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      setTokenStore(token.trim());
      await api.me();
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed");
      useAuthStore.getState().clear();
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-ink-950 px-4">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="mb-8 flex flex-col items-center justify-center gap-2 text-center">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-9 w-9 text-accent-500" />
            <span className="text-3xl font-semibold text-white tracking-tight">Custodian</span>
          </div>
          <p className="text-xs text-slate-400">Self-Hosted PaaS Control Plane</p>
        </div>

        <div className="card p-6 shadow-2xl">
          {/* First Time Setup View */}
          {setupRequired ? (
            <div>
              <div className="mb-6 border-b border-ink-800 pb-4">
                <h2 className="text-xl font-semibold text-white flex items-center gap-2">
                  <UserPlus className="h-5 w-5 text-accent-400" /> Initial Setup: Create Admin Account
                </h2>
                <p className="mt-1 text-xs text-slate-400">
                  Welcome to Custodian! Create your administrator account to get started.
                </p>
              </div>

              <form onSubmit={handleRegisterSubmit} className="space-y-4">
                <div>
                  <label className="label">Username</label>
                  <div className="relative">
                    <input
                      className="input pl-9"
                      placeholder="admin"
                      value={regUsername}
                      onChange={(e) => setRegUsername(e.target.value)}
                      required
                      autoFocus
                    />
                    <User className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
                  </div>
                </div>

                <div>
                  <label className="label">Email Address (Optional)</label>
                  <div className="relative">
                    <input
                      type="email"
                      className="input pl-9"
                      placeholder="admin@example.com"
                      value={regEmail}
                      onChange={(e) => setRegEmail(e.target.value)}
                    />
                    <Mail className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
                  </div>
                </div>

                <div>
                  <label className="label">Password</label>
                  <div className="relative">
                    <input
                      type="password"
                      className="input pl-9"
                      placeholder="At least 6 characters"
                      value={regPassword}
                      onChange={(e) => setRegPassword(e.target.value)}
                      required
                    />
                    <Lock className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
                  </div>
                </div>

                <div>
                  <label className="label">Confirm Password</label>
                  <div className="relative">
                    <input
                      type="password"
                      className="input pl-9"
                      placeholder="Re-enter password"
                      value={regConfirmPassword}
                      onChange={(e) => setRegConfirmPassword(e.target.value)}
                      required
                    />
                    <Lock className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
                  </div>
                </div>

                {error && <p className="text-xs text-rose-400 font-medium">{error}</p>}

                <button
                  type="submit"
                  className="btn-primary w-full justify-center py-2.5 mt-2"
                  disabled={busy}
                >
                  {busy ? "Creating Account..." : "Create Account & Sign In"}
                </button>
              </form>
            </div>
          ) : (
            /* Standard Login View */
            <div>
              {/* Segmented Control / Tabs */}
              <div className="mb-6 flex rounded-lg border border-ink-800 bg-ink-950 p-1 text-xs font-medium">
                <button
                  onClick={() => {
                    setTab("local");
                    setError(null);
                  }}
                  className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-md transition ${
                    tab === "local"
                      ? "bg-ink-800 text-white font-semibold shadow-sm"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  <User className="h-3.5 w-3.5" /> Local Account
                </button>
                <button
                  onClick={() => {
                    setTab("token");
                    setError(null);
                  }}
                  className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-md transition ${
                    tab === "token"
                      ? "bg-ink-800 text-white font-semibold shadow-sm"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  <KeyRound className="h-3.5 w-3.5" /> API Token
                </button>
              </div>

              {/* Tab 1: Local Password Login */}
              {tab === "local" && (
                <form onSubmit={handleLocalSubmit} className="space-y-4">
                  <div>
                    <label className="label">Username</label>
                    <div className="relative">
                      <input
                        className="input pl-9"
                        placeholder="admin"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                        autoFocus
                      />
                      <User className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
                    </div>
                  </div>
                  <div>
                    <label className="label">Password</label>
                    <div className="relative">
                      <input
                        type="password"
                        className="input pl-9"
                        placeholder="••••••••"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                      />
                      <Lock className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
                    </div>
                  </div>

                  {error && <p className="text-xs text-rose-400">{error}</p>}

                  <button
                    type="submit"
                    className="btn-primary w-full justify-center py-2.5"
                    disabled={busy}
                  >
                    {busy ? "Signing in..." : "Sign in with Local Account"}
                  </button>
                </form>
              )}

              {/* Tab 2: API Token Login */}
              {tab === "token" && (
                <form onSubmit={handleTokenSubmit} className="space-y-4">
                  <div>
                    <label className="label">Bearer API Token</label>
                    <input
                      className="input font-mono text-xs"
                      placeholder="cst_admin.xxxx..."
                      value={token}
                      onChange={(e) => setToken(e.target.value)}
                      autoFocus
                      required
                    />
                    <p className="mt-2 text-xs text-slate-500">
                      Enter an API key created via <code className="text-slate-400">custodian tokens create</code> or the dashboard.
                    </p>
                  </div>

                  {error && <p className="text-xs text-rose-400">{error}</p>}

                  <button
                    type="submit"
                    className="btn-primary w-full justify-center py-2.5"
                    disabled={busy}
                  >
                    {busy ? "Verifying Token..." : "Sign in with Token"}
                  </button>
                </form>
              )}

              {/* Optional OIDC SSO */}
              <div className="mt-6 pt-4 border-t border-ink-800">
                <a
                  href="/api/v1/auth/login"
                  className="btn-secondary w-full justify-center text-xs text-slate-400 hover:text-white"
                >
                  Single Sign-On (OIDC / OAuth2)
                </a>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
