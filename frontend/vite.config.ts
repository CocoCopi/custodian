import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The console proxies API + WebSocket traffic to the control plane during
// development so the frontend needs no CORS configuration locally.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.CUSTODIAN_API_URL ?? "http://localhost:8080",
        changeOrigin: true,
        ws: true,
      },
      "/healthz": {
        target: process.env.CUSTODIAN_API_URL ?? "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
