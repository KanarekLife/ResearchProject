import type { Group } from "./aggregate";
import type { Row } from "./games";

// Outcome colours are status colours and mean only a result.

export const OUTCOMES = [
  { key: "won", label: "won", color: "var(--good)" },
  { key: "scheme", label: "lost to the scheme", color: "var(--threat)" },
  { key: "defeated", label: "lost, hero defeated", color: "var(--hp)" },
  { key: "other", label: "error or limit", color: "var(--muted)" },
];

export function outcomeOf(r: Row): string {
  if (r.status === "won") return "won";
  if (r.status === "lost") return r.behaviour.defeated ? "defeated" : "scheme";
  return "other";
}

export function outcomeParts(g: Group) {
  const n: Record<string, number> = { won: 0, scheme: 0, defeated: 0, other: 0 };
  for (const r of g.games) n[outcomeOf(r)]++;
  return OUTCOMES.map((o) => ({ ...o, n: n[o.key] }));
}
