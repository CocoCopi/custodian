import type { ServiceStatus } from "../api/types";

const styles: Record<ServiceStatus, string> = {
  running: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
  provisioning: "bg-sky-500/10 text-sky-400 border-sky-500/30",
  deploying: "bg-amber-500/10 text-amber-400 border-amber-500/30",
  degraded: "bg-orange-500/10 text-orange-400 border-orange-500/30",
  failed: "bg-rose-500/10 text-rose-400 border-rose-500/30",
  stopped: "bg-slate-500/10 text-slate-400 border-slate-500/30",
};

export function StatusBadge({ status }: { status: ServiceStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${styles[status] ?? styles.stopped}`}
    >
      {status}
    </span>
  );
}
