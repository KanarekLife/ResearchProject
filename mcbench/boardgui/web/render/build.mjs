// POC: bundles render/entry.tsx into render/dist/render.mjs (server-side React).
import { build } from "vite";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const dir = dirname(fileURLToPath(import.meta.url));
await build({
  root: dir,
  configFile: false,
  logLevel: "warn",
  build: {
    ssr: join(dir, "entry.tsx"),
    outDir: join(dir, "dist"),
    emptyOutDir: true,
    rollupOptions: { output: { format: "es", entryFileNames: "render.mjs" } },
  },
});
