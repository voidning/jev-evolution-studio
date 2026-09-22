import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import jevEditor from "../integration/jev-vite.mjs";
export default defineConfig({
  plugins: [react(), jevEditor()],
  server: { host: "127.0.0.1" },
});
