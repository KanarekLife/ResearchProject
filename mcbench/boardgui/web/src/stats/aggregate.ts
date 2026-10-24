// Aggregates match `mcbench report` and `analyze`: means over graded games (all
// but infra_error), behaviour shares from summed counts.
import { CRITERIA, type Behaviour, type Row } from "./games";

const INFRA_ERROR = "infra_error";

export type GroupBy = "config" | "player" | "instr" | "run";
export const GROUP_BY: { key: GroupBy; label: string }[] = [
  { key: "config", label: "Player + instructions" },
  { key: "player", label: "Player" },
  { key: "instr", label: "Instructions" },
  { key: "run", label: "Run" },
];

export interface Mean {
  mean: number; // NaN with no games
  se: number;
  n: number;
}

export interface Group {
  key: string;
  player: string; // empty when the grouping leaves it out
  instr: string;
  version: string; // the instructions' hash, when the grouping keeps versions apart
  scenario: string;
  run: string;
  games: Row[];
  graded: number;
  score: Mean;
  criteria: Record<string, number>;
  status: Record<string, number>;
  roundsToWin: number;
  decisions: number;
  invalid: number;
  tokensIn: number;
  tokensOut: number;
  tokensCached: number;
  requests: number;
  ms: number;
  b: Behaviour;
}

function keyOf(
  r: Row,
  by: GroupBy,
): Omit<Group, "games" | "graded" | "score" | "criteria" | "status" | "roundsToWin" | "decisions" | "invalid" | "tokensIn" | "tokensOut" | "tokensCached" | "requests" | "ms" | "b"> {
  const none = { player: "", instr: "", version: "", run: "", scenario: r.scenario };
  const g =
    by === "config"
      ? { ...none, player: r.player, instr: r.instr, version: r.instructions_sha ?? "" }
      : by === "player"
        ? { ...none, player: r.player }
        : by === "instr"
          ? { ...none, instr: r.instr, version: r.instructions_sha ?? "" }
          : { ...none, player: r.player, instr: r.instr, run: r.run };
  return { ...g, key: [g.player, g.instr, g.version, g.run, g.scenario].join("|") };
}

export function mean(xs: number[]): Mean {
  const n = xs.length;
  if (n === 0) return { mean: NaN, se: 0, n };
  const m = xs.reduce((a, b) => a + b, 0) / n;
  if (n < 2) return { mean: m, se: 0, n };
  const v = xs.reduce((a, x) => a + (x - m) * (x - m), 0) / (n - 1);
  return { mean: m, se: Math.sqrt(v) / Math.sqrt(n), n };
}

export function emptyBehaviour(): Behaviour {
  return {
    graded: 0,
    won: 0,
    scheme: 0,
    defeated: 0,
    rounds: 0,
    basic: 0,
    attacks: 0,
    defends: 0,
    hero_defends: 0,
    ends: 0,
    ae_ends: 0,
    ae_risky: 0,
    idle_hero: 0,
    idle_ally: 0,
    card_left: 0,
    actions: 0,
    recorded: false,
  };
}

export function summarize(games: Row[]): Omit<Group, keyof ReturnType<typeof keyOf>> {
  const graded = games.filter((g) => g.status !== INFRA_ERROR);
  const status: Record<string, number> = {};
  const b = emptyBehaviour();
  for (const g of games) {
    status[g.status] = (status[g.status] ?? 0) + 1;
    for (const k of Object.keys(b) as (keyof Behaviour)[]) {
      if (k === "recorded") b.recorded ||= g.behaviour.recorded;
      else b[k] += g.behaviour[k];
    }
  }
  const criteria: Record<string, number> = {};
  for (const [c] of CRITERIA) criteria[c] = mean(graded.map((g) => g.criteria[c] ?? 0)).mean;
  const sum = (f: (g: Row) => number) => games.reduce((a, g) => a + f(g), 0);
  return {
    games,
    graded: graded.length,
    score: mean(graded.map((g) => g.score)),
    criteria,
    status,
    roundsToWin: mean(graded.filter((g) => g.status === "won").map((g) => g.rounds)).mean,
    decisions: mean(graded.map((g) => g.decisions)).mean,
    invalid: sum((g) => g.invalid),
    tokensIn: sum((g) => g.tokens_in),
    tokensOut: sum((g) => g.tokens_out),
    tokensCached: sum((g) => g.tokens_cached ?? 0),
    requests: sum((g) => g.requests ?? 0),
    ms: sum((g) => g.ms),
    b,
  };
}

export function group(rows: Row[], by: GroupBy): Group[] {
  const groups = new Map<string, { k: ReturnType<typeof keyOf>; games: Row[] }>();
  for (const r of rows) {
    const k = keyOf(r, by);
    const g = groups.get(k.key) ?? { k, games: [] };
    g.games.push(r);
    groups.set(k.key, g);
  }
  return [...groups.values()].map(({ k, games }) => ({ ...k, ...summarize(games) }));
}

export const ratio = (n: number, d: number) => (d ? n / d : NaN);

// days buckets the rows by their date in date order, each summarized the way
// the KPI total is; the KPI sparklines draw one value per day from it.
export function days(rows: Row[]): { date: string; s: ReturnType<typeof summarize> }[] {
  const by = new Map<string, Row[]>();
  for (const r of rows) if (r.date) by.set(r.date, [...(by.get(r.date) ?? []), r]);
  return [...by]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, rs]) => ({ date, s: summarize(rs) }));
}
