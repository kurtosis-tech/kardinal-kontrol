import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import checker from "vite-plugin-checker";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const isDev = mode === "development";
  return {
    plugins: [react(), ...(isDev ? [checker({ typescript: true })] : [])],
    resolve: {
      alias: {
        "@": "/src",
      },
    },
    server: {
      proxy: {
        // local dev mode proxy, default is local kontrol service.
        // change this to hit production etc if desired.
        // in production, this behavior is handled by kardinal nginx ingress
        "/api": {
          // target: "https://app.kardinal.dev",
          target: "http://localhost:8080",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ""),
        },
      },
    },
  };
});
