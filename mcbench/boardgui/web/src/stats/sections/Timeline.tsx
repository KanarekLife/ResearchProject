import { useMemo } from "react";
import { parseRun } from "../../runs";
import { summarize } from "../aggregate";
import type { Row } from "../games";
import { DAY, fixed, pct } from "../format";
import { useMounted } from "../motion";
import { useTip } from "../tooltip";

export function Timeline({ rows }: { rows: Row[] }) {
  const tip = useTip();
  const on = useMounted();
  const runs = useMemo(() => {
    const by = new Map<string, Row[]>();
    for (const r of rows) by.set(r.run, [...(by.get(r.run) ?? []), r]);
    return [...by].map(([run, games]) => ({ run, name: parseRun(run), t: runTime(run), s: summarize(games) })).filter((r) => Number.isFinite(r.t));
  }, [rows]);
  if (runs.length === 0) return <p className="dim">No run has a date in its name.</p>;
  const W = 1000,
    H = 220,
    P = { l: 36, r: 12, t: 10, b: 26 };
  let t0 = Math.min(...runs.map((r) => r.t)),
    t1 = Math.max(...runs.map((r) => r.t));
  if (t1 - t0 < DAY) [t0, t1] = [t0 - DAY / 2, t1 + DAY / 2];
  const x = (t: number) => P.l + ((t - t0) / (t1 - t0)) * (W - P.l - P.r);
  const y = (v: number) => P.t + (1 - v) * (H - P.t - P.b);
  const days: number[] = [];
  for (let d = Math.ceil(t0 / DAY) * DAY; d <= t1; d += DAY) days.push(d);
  const step = Math.ceil(days.length / 10);
  return (
    <svg className="timeline" viewBox={`0 0 ${W} ${H}`} role="img" aria-label="Mean score of each run over time">
      {[0, 0.25, 0.5, 0.75, 1].map((v) => (
        <g key={v}>
          <line x1={P.l} x2={W - P.r} y1={y(v)} y2={y(v)} className="grid" />
          <text x={P.l - 6} y={y(v) + 4} textAnchor="end" className="axis">
            {v}
          </text>
        </g>
      ))}
      {days
        .filter((_, i) => i % step === 0)
        .map((d) => (
          <text key={d} x={x(d)} y={H - 6} textAnchor="middle" className="axis">
            {new Date(d).toISOString().slice(5, 10)}
          </text>
        ))}
      {runs.map((r, i) => (
        <a key={r.run} href={"./?" + new URLSearchParams({ run: r.run })}>
          <circle
            cx={x(r.t)}
            cy={on ? y(r.s.score.mean || 0) : y(0)}
            r={on ? 4 + Math.sqrt(r.s.games.length) * 1.4 : 0}
            className="run-dot"
            style={{ transitionDelay: `${i * 25}ms` }}
            tabIndex={0}
            {...tip(() => (
              <div className="gtip">
                <div>
                  <b>{fixed(3)(r.s.score.mean)}</b> mean score · {pct(r.s.criteria.win)} won
                </div>
                <div className="dim">
                  {r.name.player.replace(/^model:/, "")} · {r.name.instructions || "(none)"}
                </div>
                <div className="dim">
                  {r.name.date} {r.name.time} · {r.s.games.length} games
                </div>
              </div>
            ))}
          />
        </a>
      ))}
    </svg>
  );
}

function runTime(run: string): number {
  const r = parseRun(run);
  return r.date ? Date.parse(`${r.date}T${r.time}:00Z`) : NaN;
}
