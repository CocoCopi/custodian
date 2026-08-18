import { useEffect, useRef } from "react";
import type { LogEntry } from "../api/types";

export function LogViewer({ entries, connected }: { entries: LogEntry[]; connected: boolean }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    ref.current?.scrollTo({ top: ref.current.scrollHeight });
  }, [entries]);

  return (
    <div className="card flex h-96 flex-col overflow-hidden">
      <div className="flex items-center justify-between border-b border-ink-800 px-4 py-2">
        <span className="text-xs font-medium uppercase tracking-wide text-slate-400">
          Live logs
        </span>
        <span
          className={`flex items-center gap-1.5 text-xs ${
            connected ? "text-emerald-400" : "text-slate-500"
          }`}
        >
          <span
            className={`h-2 w-2 rounded-full ${connected ? "bg-emerald-400" : "bg-slate-600"}`}
          />
          {connected ? "streaming" : "disconnected"}
        </span>
      </div>
      <div ref={ref} className="flex-1 overflow-y-auto bg-black/40 p-4 font-mono text-xs leading-5">
        {entries.length === 0 ? (
          <p className="text-slate-600">Waiting for log output…</p>
        ) : (
          entries.map((entry, i) => (
            <div
              key={i}
              className={`whitespace-pre-wrap ${
                entry.stream === "stderr" ? "text-rose-400" : "text-slate-300"
              }`}
            >
              {entry.message}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
