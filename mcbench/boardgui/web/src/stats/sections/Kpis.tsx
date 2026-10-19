import { ratio, type summarize } from "../aggregate";
import { compact, duration, fixed, pct, thousands } from "../format";
import { AnimatedNumber } from "../marks";

export function Kpis({ total, runs }: { total: ReturnType<typeof summarize>; runs: number }) {
  const n = total.games.length;
  const tiles: { label: string; value: number; format: (x: number) => string; sub: string }[] = [
    { label: "Games", value: n, format: thousands, sub: `${runs} run${runs === 1 ? "" : "s"}` },
    { label: "Win rate", value: total.criteria.win, format: pct, sub: `${total.status.won ?? 0} won of ${total.graded} graded` },
    { label: "Mean score", value: total.score.mean, format: fixed(3), sub: `± ${total.score.se.toFixed(3)} s.e.` },
    { label: "Rounds to win", value: total.roundsToWin, format: fixed(1), sub: `${fixed(0)(total.decisions)} decisions per game` },
    { label: "Invalid choices", value: total.invalid, format: thousands, sub: `${fixed(2)(ratio(total.invalid, n))} per game` },
    { label: "Tokens per game", value: total.tokensIn / n, format: compact, sub: `${compact(total.tokensOut / n)} out · ${duration(total.ms / n)}` },
  ];
  return (
    <div className="kpis">
      {tiles.map((t, i) => (
        <div className="kpi panel" key={t.label} style={{ animationDelay: `${i * 50}ms` }}>
          <div className="kpi-label">{t.label}</div>
          <div className="kpi-value">
            <AnimatedNumber value={t.value} format={t.format} />
          </div>
          <div className="kpi-sub dim">{t.sub}</div>
        </div>
      ))}
    </div>
  );
}
