import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// During development the React app calls /api/* which Vite proxies to the Go
// server on :8080, so there are no CORS issues and no hard-coded host.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.API_PROXY_TARGET ?? "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
