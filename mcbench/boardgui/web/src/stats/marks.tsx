import { pct } from "./format";
import { useCountUp, useMounted } from "./motion";
import type { useTip } from "./tooltip";

export function Meter({ value, se, color = "var(--series-1)", label }: { value: number; se?: number; color?: string; label?: string }) {
  const on = useMounted();
  const v = Number.isFinite(value) ? Math.max(0, Math.min(1, value)) : 0;
  return (
    <div className="meter">
      <div className="meter-track">
        <div className="meter-fill" style={{ width: on ? `${v * 100}%` : 0, background: color }} />
        {se ? <div className="meter-se" style={{ left: `${Math.max(0, v - se) * 100}%`, width: on ? `${Math.min(2 * se, 1) * 100}%` : 0 }} /> : null}
      </div>
      {label !== undefined && <span className="meter-label">{label}</span>}
    </div>
  );
}

export function Stack({ parts, tip }: { parts: { key: string; label: string; n: number; color: string }[]; tip: ReturnType<typeof useTip> }) {
  const on = useMounted();
  const total = parts.reduce((a, p) => a + p.n, 0);
  return (
    <div className="stack">
      {parts
        .filter((p) => p.n > 0)
        .map((p) => (
          <div
            key={p.key}
            className="stack-part"
            tabIndex={0}
            style={{ flexGrow: on ? p.n : 0, background: p.color }}
            {...tip(() => (
              <>
                <b>{p.n}</b> {p.label} <span className="dim">· {pct(p.n / total)}</span>
              </>
            ))}
          />
        ))}
    </div>
  );
}

export function AnimatedNumber({ value, format }: { value: number; format: (n: number) => string }) {
  const v = useCountUp(value);
  return <>{Number.isFinite(value) ? format(v) : "–"}</>;
}

// Spark is the small line chart of a value's day-by-day history beside a KPI.
// The viewBox is stretched to the box, so strokes do not scale with it.
export function Spark({ values, color = "var(--series-1)" }: { values: number[]; color?: string }) {
  const on = useMounted();
  const finite = values.filter(Number.isFinite);
  if (finite.length < 2) return null;
  const min = Math.min(...finite),
    max = Math.max(...finite);
  const y = (v: number) => 30 - 27 * (max > min ? (v - min) / (max - min) : 0.5) - 1.5;
  let d = "",
    pen = false;
  values.forEach((v, i) => {
    if (!Number.isFinite(v)) {
      pen = false;
      return;
    }
    d += `${pen ? "L" : "M"}${(((i + 0.5) / values.length) * 100).toFixed(1)},${y(v).toFixed(1)} `;
    pen = true;
  });
  return (
    <svg
      className="spark"
      viewBox="0 0 100 30"
      preserveAspectRatio="none"
      role="img"
      aria-label="Value over time"
      style={{ opacity: on ? 1 : 0, transition: "opacity .6s var(--ease, .3s)" }}
    >
      <path d={d} fill="none" stroke={color} strokeWidth={1.5} vectorEffect="non-scaling-stroke" />
    </svg>
  );
}
