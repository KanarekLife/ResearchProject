import { useMemo } from "react";
import { days, ratio, type summarize } from "../aggregate";
import type { Row } from "../games";
import { compact, duration, fixed, pct, thousands } from "../format";
import { AnimatedNumber, Spark } from "../marks";

type Day = ReturnType<typeof summarize>;

export function Kpis({ total, runs, rows }: { total: ReturnType<typeof summarize>; runs: number; rows: Row[] }) {
  const n = total.games.length;
  const perDay = useMemo(() => days(rows), [rows]);
  // over gives each tile's value day by day, in date order, for its sparkline.
  const over = (f: (s: Day) => number) => perDay.map((d) => f(d.s));
  const tiles: { label: string; value: number; format: (x: number) => string; sub: string; spark: number[] }[] = [
    { label: "Games", value: n, format: thousands, sub: `${runs} run${runs === 1 ? "" : "s"}`, spark: over((s) => s.games.length) },
    { label: "Win rate", value: total.criteria.win, format: pct, sub: `${total.status.won ?? 0} won of ${total.graded} graded`, spark: over((s) => s.criteria.win) },
    { label: "Mean score", value: total.score.mean, format: fixed(3), sub: `± ${total.score.se.toFixed(3)} s.e.`, spark: over((s) => s.score.mean) },
    { label: "Rounds to win", value: total.roundsToWin, format: fixed(1), sub: `${fixed(0)(total.decisions)} decisions per game`, spark: over((s) => s.roundsToWin) },
    { label: "Invalid choices", value: total.invalid, format: thousands, sub: `${fixed(2)(ratio(total.invalid, n))} per game`, spark: over((s) => s.invalid) },
    { label: "Tokens per game", value: total.tokensIn / n, format: compact, sub: `${compact(total.tokensOut / n)} out · ${duration(total.ms / n)}`, spark: over((s) => s.tokensIn / s.games.length) },
    { label: "Requests per game", value: total.requests / n, format: fixed(1), sub: `${thousands(total.requests)} total`, spark: over((s) => s.requests / s.games.length) },
    { label: "Total tokens", value: total.tokensIn + total.tokensOut, format: compact, sub: `${compact(total.tokensIn)} in · ${compact(total.tokensOut)} out${total.tokensCached ? ` · ${compact(total.tokensCached)} cached` : ""}`, spark: over((s) => s.tokensIn + s.tokensOut) },
  ];
  return (
    <div className="kpis">
      {tiles.map((t, i) => (
        <div className="kpi panel" key={t.label} style={{ animationDelay: `${i * 50}ms` }}>
          <div className="kpi-label">{t.label}</div>
          <div className="kpi-value">
            <AnimatedNumber value={t.value} format={t.format} />
            <Spark values={t.spark} />
          </div>
          <div className="kpi-sub dim">{t.sub}</div>
        </div>
      ))}
    </div>
  );
}
