// A game as benchmark.GameStats exports it.
import { parseRun } from "../runs";

export interface Behaviour {
  graded: number;
  won: number;
  scheme: number;
  defeated: number;
  rounds: number;
  basic: number;
  attacks: number;
  defends: number;
  hero_defends: number;
  ends: number;
  ae_ends: number;
  ae_risky: number;
  idle_hero: number;
  idle_ally: number;
  card_left: number;
  actions: number;
  recorded: boolean;
}

export interface Game {
  run: string;
  game: string;
  scenario: string;
  player: string;
  instructions: string;
  instructions_sha?: string;
  seed: number;
  sample: number;
  status: string;
  error?: string;
  score: number;
  criteria: Record<string, number>;
  rounds: number;
  decisions: number;
  invalid: number;
  stats: Record<string, number>;
  tokens_in: number;
  tokens_out: number;
  ms: number;
  behaviour: Behaviour;
}

export interface Row extends Game {
  date: string;
  time: string;
  model: string;
  instr: string; // the instruction names, "(none)" for none
}

export const CRITERIA: [string, string][] = [
  ["win", "win"],
  ["villain_damage", "villain dmg"],
  ["hero_hp", "hero hp"],
  ["threat", "threat"],
  ["speed", "speed"],
];

export function modelOf(player: string): string {
  return player.replace(/^model:/, "").split("@")[0];
}

export function toRows(games: Game[]): Row[] {
  return games.map((g) => {
    const r = parseRun(g.run);
    return { ...g, date: r.date, time: r.time, model: modelOf(g.player), instr: g.instructions || "(none)" };
  });
}
