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
