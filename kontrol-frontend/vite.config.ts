import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import checker from "vite-plugin-checker";

// https://vitejs.dev/config/
export default defineConfig({
  // TODO: Checker makes the nix build halt at the end: (buildPhase): ✓ built in 2.34s
  // plugins: [react(), checker({ typescript: true })],
  plugins: [react()],
  resolve: {
    alias: {
      "@": "/src",
    },
  },
});
