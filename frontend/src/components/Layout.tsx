import { Activity, Boxes, Key, LayoutDashboard, LogOut, ShieldCheck } from "lucide-react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuthStore } from "../store/auth";

const nav = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard, end: true },
  { to: "/apps", label: "Applications", icon: Boxes },
  { to: "/tokens", label: "API Tokens", icon: Key },
  { to: "/health", label: "System Health", icon: Activity },
];

export function Layout() {
  const clear = useAuthStore((s) => s.clear);
  const navigate = useNavigate();

  return (
    <div className="flex min-h-screen">
      <aside className="flex w-60 flex-col border-r border-ink-800 bg-ink-900 px-4 py-6">
        <div className="mb-8 flex items-center gap-2 px-2">
          <ShieldCheck className="h-6 w-6 text-accent-500" />
          <span className="text-lg font-semibold text-white">Custodian</span>
        </div>
        <nav className="flex flex-1 flex-col gap-1">
          {nav.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-ink-800 text-accent-400"
                    : "text-slate-400 hover:bg-ink-800 hover:text-slate-200"
                }`
              }
            >
              <Icon className="h-4 w-4" />
              {label}
            </NavLink>
          ))}
        </nav>
        <button
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-400 hover:bg-ink-800 hover:text-slate-200"
          onClick={() => {
            clear();
            navigate("/login");
          }}
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </button>
      </aside>
      <main className="flex-1 overflow-y-auto p-8">
        <Outlet />
      </main>
    </div>
  );
}
