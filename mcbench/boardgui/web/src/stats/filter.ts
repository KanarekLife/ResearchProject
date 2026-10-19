import type { Row } from "./games";

export interface Filters {
  from: string;
  to: string;
  model: string[];
  instr: string[];
  scenario: string[];
  status: string[];
}

export type Facet = "model" | "instr" | "scenario" | "status";
export const FACETS: { key: Facet; label: string }[] = [
  { key: "model", label: "Player" },
  { key: "instr", label: "Instructions" },
  { key: "scenario", label: "Scenario" },
  { key: "status", label: "Result" },
];

export function matches(r: Row, f: Filters, skip?: Facet): boolean {
  if (f.from && r.date < f.from) return false;
  if (f.to && r.date > f.to) return false;
  return FACETS.every(({ key }) => key === skip || f[key].length === 0 || f[key].includes(r[key]));
}

// facetCounts ignores the facet's own filter, so each option shows what it would add.
export function facetCounts(rows: Row[], f: Filters, facet: Facet): Map<string, number> {
  const out = new Map<string, number>();
  for (const r of rows) if (matches(r, f, facet)) out.set(r[facet], (out.get(r[facet]) ?? 0) + 1);
  return out;
}
