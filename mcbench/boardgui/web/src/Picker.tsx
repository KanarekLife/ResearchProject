import { useMemo, useState } from "react";
import { byNumber, parseRun, type RunName } from "./runs";
import type { Runs } from "./types";

export interface Selection {
  run: string;
  game: string;
}

function counted(values: string[]): [string, number][] {
  const n = new Map<string, number>();
  for (const v of values) n.set(v, (n.get(v) ?? 0) + 1);
  return [...n];
}

export function Picker({ runs, sel, following, choose }: {
  runs: Runs | null;
  sel: Selection | null;
  following?: string; // the game shown while following the run
  choose: (s: Selection) => void;
}) {
  const all = useMemo(() => (runs?.runs ?? []).map((r) => parseRun(r.id)), [runs]);
  const [date, setDate] = useState(""); // empty: any
  const [player, setPlayer] = useState("");

  const onDate = all.filter((r) => !date || r.date === date);
  const shown = onDate.filter((r) => !player || r.player === player);
  const cur = all.find((r) => r.id === sel?.run);
  const run = runs?.runs.find((r) => r.id === sel?.run);
  const games = [...(run?.games ?? [])].sort((a, b) => byNumber(a.name, b.name));
  // Only a local run can still be played, so only it can be followed.
  const game = sel?.game || (run?.local ? "" : (following ?? ""));

  // A narrower filter that leaves out the open run opens its newest run.
  const narrow = (d: string, p: string) => {
    setDate(d);
    setPlayer(p);
    const match = all.find((r) => (!d || r.date === d) && (!p || r.player === p));
    if (match && !(cur && (!d || cur.date === d) && (!p || cur.player === p))) choose({ run: match.id, game: "" });
  };

  return (
    <>
      <select title="Date" value={date} onChange={(e) => narrow(e.target.value, "")}>
        <option value="">any date</option>
        {counted(all.map((r) => r.date)).map(([d, n]) => <option key={d} value={d}>{`${d || "undated"} (${n})`}</option>)}
      </select>
      <select title="Player" value={player} onChange={(e) => narrow(date, e.target.value)}>
        <option value="">any player</option>
        {counted(onDate.map((r) => r.player)).sort((a, b) => byNumber(a[0], b[0])).map(([p, n]) => <option key={p} value={p}>{`${p} (${n})`}</option>)}
      </select>
      <select title="Run" className="wide" value={sel?.run ?? ""} onChange={(e) => choose({ run: e.target.value, game: "" })}>
        {sel && !shown.some((r) => r.id === sel.run) && <option value={sel.run}>{runLabel(cur ?? parseRun(sel.run), date)}</option>}
        {shown.map((r) => <option key={r.id} value={r.id}>{runLabel(r, date)}</option>)}
      </select>
      <select title="Game" value={game} onChange={(e) => choose({ run: sel!.run, game: e.target.value })}>
        {run?.local && <option value="">{"follow the run" + (following && !sel?.game ? ` (${following})` : "")}</option>}
        {games.map((g) => <option key={g.name} value={g.name}>{`${g.name} · ${g.state}`}</option>)}
      </select>
    </>
  );
}

function runLabel(r: RunName, date: string): string {
  const when = date ? r.time : `${r.date} ${r.time}`;
  return [when.trim(), r.player, r.instructions].filter(Boolean).join(" · ");
}
