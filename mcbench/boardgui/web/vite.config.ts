import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The Go server embeds ../static, so asset URLs are relative and file names are
// fixed: a rebuild overwrites the same files and keeps the committed placeholder.
export default defineConfig({
  base: "./",
  plugins: [react()],
  build: {
    outDir: "../static",
    emptyOutDir: false,
    rolldownOptions: {
      input: { board: "index.html", stats: "stats.html" },
      output: { entryFileNames: "assets/[name].js", chunkFileNames: "assets/[name].js", assetFileNames: "assets/[name][extname]" },
    },
  },
  // `npm run dev` asks a running `go run ./cmd board_gui` for the data.
  server: { proxy: { "/api": "http://127.0.0.1:8090" } },
});
