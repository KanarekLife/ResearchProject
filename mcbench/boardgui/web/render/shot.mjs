// POC: reads {"view", "codes"} JSON on stdin and writes a PNG of the board to argv[2].
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { renderBoard } from "./dist/render.mjs";

const dir = dirname(fileURLToPath(import.meta.url));
const css = readFileSync(join(dir, "../src/app.css"), "utf8");
const { view, codes } = JSON.parse(readFileSync(0, "utf8"));
const out = process.argv[2];

const page = mkdtempSync(join(tmpdir(), "board-"));
const html = join(page, "board.html");
writeFileSync(html, `<!doctype html><html><head><meta charset="utf-8"><style>${css}</style></head><body>${renderBoard(view, codes)}</body></html>`);
execFileSync(process.env.CHROMIUM || "chromium-browser", [
  "--headless", "--no-sandbox", "--disable-gpu", "--hide-scrollbars",
  "--window-size=900,950", "--virtual-time-budget=4000",
  `--screenshot=${out}`, "file://" + html,
], { stdio: ["ignore", "ignore", "inherit"] });
