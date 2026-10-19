import type { Filters } from "./filter";
import { DAY } from "./format";

export function DateRange({ f, set, dates }: { f: Filters; set: (f: Filters) => void; dates: string[] }) {
  const last = dates[dates.length - 1];
  // Presets count back from the newest run, not today.
  const since = (days: number) => (last ? new Date(Date.parse(last) - (days - 1) * DAY).toISOString().slice(0, 10) : "");
  const presets: [string, string][] = [
    ["All", ""],
    ["Last day", since(1)],
    ["7 days", since(7)],
    ["30 days", since(30)],
  ];
  return (
    <div className="dates">
      <div className="seg">
        {presets.map(([label, from]) => (
          <button key={label} className={f.from === from && !f.to ? "on" : ""} onClick={() => set({ ...f, from, to: "" })}>
            {label}
          </button>
        ))}
      </div>
      <select aria-label="From date" value={f.from} onChange={(e) => set({ ...f, from: e.target.value })}>
        <option value="">from start</option>
        {dates.map((d) => (
          <option key={d} value={d}>
            {d}
          </option>
        ))}
      </select>
      <span className="dim">→</span>
      <select aria-label="To date" value={f.to} onChange={(e) => set({ ...f, to: e.target.value })}>
        <option value="">to latest</option>
        {dates.map((d) => (
          <option key={d} value={d}>
            {d}
          </option>
        ))}
      </select>
    </div>
  );
}
