import { useEffect } from "react";
import { useLogStore } from "../store/logs";

/**
 * Opens a WebSocket to the control plane and streams log entries into the
 * log store. Automatically reconnects while `active` is true.
 */
export function useLogStream(serviceId: string | undefined, active: boolean) {
  const append = useLogStore((s) => s.append);
  const setConnected = useLogStore((s) => s.setConnected);
  const clear = useLogStore((s) => s.clear);

  useEffect(() => {
    if (!serviceId || !active) return;

    const token = localStorage.getItem("custodian_token");
    const proto = window.location.protocol === "https:" ? "wss" : "ws";
    const base = import.meta.env.VITE_API_URL ?? window.location.host;
    const url = `${proto}://${base}/api/v1/services/${serviceId}/logs?token=${encodeURIComponent(token ?? "")}`;

    let socket: WebSocket | null = null;
    let retry: ReturnType<typeof setTimeout> | null = null;
    let closed = false;

    const connect = () => {
      socket = new WebSocket(url);
      socket.onopen = () => setConnected(true);
      socket.onmessage = (event) => {
        try {
          append(JSON.parse(event.data));
        } catch {
          // ignore malformed frames
        }
      };
      socket.onclose = () => {
        setConnected(false);
        if (!closed) retry = setTimeout(connect, 2000);
      };
    };

    clear();
    connect();
    return () => {
      closed = true;
      if (retry) clearTimeout(retry);
      socket?.close();
      setConnected(false);
    };
  }, [serviceId, active, append, clear, setConnected]);
}
