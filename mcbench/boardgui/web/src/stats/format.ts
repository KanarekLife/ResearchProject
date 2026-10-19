export const DAY = 864e5;

export const pct = (x: number) => (Number.isFinite(x) ? `${Math.round(x * 100)}%` : "–");
export const fixed = (d: number) => (x: number) => (Number.isFinite(x) ? x.toFixed(d) : "–");
export const thousands = (x: number) => (Number.isFinite(x) ? Math.round(x).toLocaleString("en-US") : "–");

export function compact(x: number): string {
  if (!Number.isFinite(x) || x === 0) return "–";
  if (x >= 1e9) return (x / 1e9).toFixed(1) + "B";
  if (x >= 1e6) return (x / 1e6).toFixed(1) + "M";
  if (x >= 1e3) return (x / 1e3).toFixed(1) + "k";
  return String(Math.round(x));
}

export function duration(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) return "–";
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m${String(s % 60).padStart(2, "0")}s`;
  return `${Math.floor(m / 60)}h${String(m % 60).padStart(2, "0")}m`;
}
