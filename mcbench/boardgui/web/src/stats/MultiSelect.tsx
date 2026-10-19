import { useEffect, useRef, useState } from "react";
import { byNumber } from "../runs";
import { facetCounts, type Facet, type Filters } from "./filter";
import type { Row } from "./games";

export function MultiSelect({ label, facet, rows, f, set }: { label: string; facet: Facet; rows: Row[]; f: Filters; set: (f: Filters) => void }) {
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState("");
  const box = useRef<HTMLDivElement>(null);
  const search = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (!open) return;
    search.current?.focus();
    const close = (e: Event) => {
      if (e instanceof KeyboardEvent ? e.key === "Escape" : !box.current?.contains(e.target as Node)) setOpen(false);
    };
    addEventListener("pointerdown", close);
    addEventListener("keydown", close);
    return () => {
      removeEventListener("pointerdown", close);
      removeEventListener("keydown", close);
    };
  }, [open]);

  const chosen = f[facet];
  const counts = facetCounts(rows, f, facet);
  for (const v of chosen) if (!counts.has(v)) counts.set(v, 0);
  const values = [...counts].filter(([v]) => v.toLowerCase().includes(q.toLowerCase())).sort((a, b) => byNumber(a[0], b[0]));
  const toggle = (v: string) => set({ ...f, [facet]: chosen.includes(v) ? chosen.filter((x) => x !== v) : [...chosen, v] });
  const summary = chosen.length === 0 ? "all" : chosen.length === 1 ? chosen[0] : `${chosen.length} selected`;

  return (
    <div className="multi" ref={box}>
      <button className={"multi-btn" + (chosen.length ? " active" : "")} aria-expanded={open} onClick={() => setOpen(!open)}>
        <span className="dim">{label}</span> <span className="multi-sum">{summary}</span> <span className="caret">▾</span>
      </button>
      <div className={"pop" + (open ? " open" : "")} aria-hidden={!open}>
        <input ref={search} className="search" placeholder={`Search ${label.toLowerCase()}…`} value={q} onChange={(e) => setQ(e.target.value)} />
        <ul>
          {values.map(([v, n]) => (
            <li key={v}>
              <label className={n === 0 && !chosen.includes(v) ? "none" : undefined}>
                <input type="checkbox" checked={chosen.includes(v)} onChange={() => toggle(v)} />
                <span className="v">{v}</span>
                <span className="n">{n}</span>
              </label>
            </li>
          ))}
          {values.length === 0 && <li className="dim empty-li">nothing matches</li>}
        </ul>
        <div className="pop-foot">
          <button onClick={() => set({ ...f, [facet]: [] })} disabled={!chosen.length}>
            Clear
          </button>
          <button onClick={() => set({ ...f, [facet]: values.filter(([, n]) => n > 0).map(([v]) => v) })}>Select shown</button>
        </div>
      </div>
    </div>
  );
}
