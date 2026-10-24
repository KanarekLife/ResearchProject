import { useMemo, useState, type ReactNode } from "react";
import { boardLink, byNumber } from "../../runs";
import type { Row } from "../games";
import { compact, duration, thousands } from "../format";
import { Meter } from "../marks";
import { outcomeOf } from "../outcomes";

type Col = { label: string; sort: (r: Row) => number | string; cell: (r: Row) => ReactNode; num?: boolean };

const frac = (r: Row, a: string, b: string) => `${r.stats?.[a] ?? 0}/${r.stats?.[b] ?? 0}`;

const COLS: Col[] = [
  {
    label: "When",
    sort: (r) => r.run,
    cell: (r) => (
      <span className="dim">
        {r.date} {r.time}
      </span>
    ),
  },
  { label: "Player", sort: (r) => r.player, cell: (r) => r.player.replace(/^model:/, "") },
  { label: "Instructions", sort: (r) => r.instr, cell: (r) => <span className="dim">{r.instr}</span> },
  { label: "Seed", sort: (r) => r.seed * 1000 + r.sample, cell: (r) => r.seed + (r.sample ? `/s${r.sample}` : ""), num: true },
  {
    label: "Result",
    sort: (r) => r.status,
    cell: (r) => (
      <span className={"res " + outcomeOf(r)} title={r.error}>
        {r.status}
      </span>
    ),
  },
  { label: "Score", sort: (r) => r.score, cell: (r) => <Meter value={r.score} label={r.score.toFixed(3)} /> },
  { label: "Rounds", sort: (r) => r.rounds, cell: (r) => r.rounds, num: true },
  { label: "Villain", sort: (r) => r.criteria.villain_damage ?? 0, cell: (r) => frac(r, "villain_damage", "villain_total_hp"), num: true },
  { label: "Hero HP", sort: (r) => r.criteria.hero_hp ?? 0, cell: (r) => frac(r, "hero_hp", "hero_max_hp"), num: true },
  { label: "Threat", sort: (r) => r.stats?.main_threat ?? 0, cell: (r) => frac(r, "main_threat", "main_target"), num: true },
  { label: "Invalid", sort: (r) => r.invalid, cell: (r) => r.invalid, num: true },
  { label: "Tokens", sort: (r) => r.tokens_in + r.tokens_out, cell: (r) => (
    <span title={`${thousands(r.tokens_in)} in · ${thousands(r.tokens_out)} out${r.tokens_cached ? ` · ${compact(r.tokens_cached)} cached` : ""}`}>
      {compact(r.tokens_in + r.tokens_out)}
    </span>
  ), num: true },
  { label: "Time", sort: (r) => r.ms, cell: (r) => duration(r.ms), num: true },
];

const PAGE = 25;

export function GamesTable({ rows }: { rows: Row[] }) {
  const [sort, setSort] = useState({ col: 0, desc: true });
  const [shown, setShown] = useState(PAGE);
  const sorted = useMemo(() => {
    const k = COLS[sort.col].sort;
    return [...rows].sort((a, b) => {
      const x = k(a),
        y = k(b);
      const c = typeof x === "number" && typeof y === "number" ? x - y : byNumber(String(x), String(y));
      return sort.desc ? -c : c;
    });
  }, [rows, sort]);
  return (
    <>
      <div className="scroll-x">
        <table className="games">
          <thead>
            <tr>
              {COLS.map((c, i) => (
                <th
                  key={c.label}
                  className={(c.num ? "num " : "") + "sortable"}
                  aria-sort={sort.col === i ? (sort.desc ? "descending" : "ascending") : "none"}
                  onClick={() => setSort({ col: i, desc: sort.col === i ? !sort.desc : true })}
                >
                  {c.label}
                  {sort.col === i ? (sort.desc ? " ↓" : " ↑") : ""}
                </th>
              ))}
              <th />
            </tr>
          </thead>
          <tbody>
            {sorted.slice(0, shown).map((r) => (
              <tr key={r.run + r.game}>
                {COLS.map((c) => (
                  <td key={c.label} className={c.num ? "num" : undefined}>
                    {c.cell(r)}
                  </td>
                ))}
                <td>
                  <a className="replay" href={boardLink(r.run, r.game)}>
                    Replay ▸
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {shown < sorted.length && (
        <button className="more" onClick={() => setShown(shown + PAGE * 2)}>
          Show more · {sorted.length - shown} left
        </button>
      )}
    </>
  );
}
