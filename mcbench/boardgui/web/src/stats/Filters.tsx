import { GROUP_BY, type GroupBy } from "./aggregate";
import { FACETS, type Filters } from "./filter";
import type { Row } from "./games";
import { DateRange } from "./DateRange";
import { MultiSelect } from "./MultiSelect";

export function FilterBar({
  rows,
  f,
  set,
  by,
  setBy,
  dates,
}: {
  rows: Row[];
  f: Filters;
  set: (f: Filters) => void;
  by: GroupBy;
  setBy: (b: GroupBy) => void;
  dates: string[]; // every run date, ascending
}) {
  const active = !!(f.from || f.to || FACETS.some(({ key }) => f[key].length));
  return (
    <div className="filters">
      <DateRange f={f} set={set} dates={dates} />
      {FACETS.map(({ key, label }) => (
        <MultiSelect key={key} label={label} facet={key} rows={rows} f={f} set={set} />
      ))}
      <div className="seg" role="radiogroup" aria-label="Group by">
        <span className="seg-label">Group by</span>
        {GROUP_BY.map((g) => (
          <button key={g.key} role="radio" aria-checked={by === g.key} className={by === g.key ? "on" : ""} onClick={() => setBy(g.key)}>
            {g.label}
          </button>
        ))}
      </div>
      <button className={"reset" + (active ? "" : " hidden")} onClick={() => set({ from: "", to: "", model: [], instr: [], scenario: [], status: [] })}>
        Clear filters
      </button>
    </div>
  );
}
