// The statistics page; its filters live in the address so a view can be shared.
import { useEffect, useMemo, useState } from "react";
import { Nav } from "../Nav";
import { fromAddress, toAddress } from "./address";
import { group, summarize, type Group } from "./aggregate";
import { matches } from "./filter";
import { toRows, type Game, type Row } from "./games";
import { FilterBar } from "./Filters";
import { Versioned, versioned } from "./names";
import { Section } from "./Section";
import { BehaviourTable } from "./sections/Behaviour";
import { Criteria } from "./sections/Criteria";
import { GamesTable } from "./sections/GamesTable";
import { Kpis } from "./sections/Kpis";
import { Leaderboard, SortPicker, type SortKey } from "./sections/Leaderboard";
import { Strip } from "./sections/Strip";
import { Timeline } from "./sections/Timeline";
import { TooltipLayer } from "./tooltip";

export function Stats() {
  const [all, setAll] = useState<Row[] | null>(null);
  const [error, setError] = useState("");
  const [{ f, by }, setView] = useState(fromAddress);
  const [sort, setSort] = useState<SortKey>("score");

  useEffect(() => {
    fetch("api/stats")
      .then((r) => (r.ok ? r.json() : r.text().then((t) => Promise.reject(new Error(t)))))
      .then((g: Game[]) => setAll(toRows(g)))
      .catch((e) => setError(String(e.message ?? e)));
  }, []);
  useEffect(() => toAddress(f, by), [f, by]);

  const rows = useMemo(() => (all ?? []).filter((r) => matches(r, f)), [all, f]);
  const dates = useMemo(() => [...new Set((all ?? []).map((r) => r.date).filter(Boolean))].sort(), [all]);
  const groups = useMemo(() => {
    const gs = group(rows, by);
    const key: Record<SortKey, (g: Group) => number> = {
      score: (g) => g.score.mean,
      win: (g) => g.criteria.win,
      games: (g) => g.games.length,
      tokens: (g) => g.tokensIn / g.games.length,
    };
    return gs.sort((a, b) => (key[sort](b) || 0) - (key[sort](a) || 0) || a.key.localeCompare(b.key));
  }, [rows, by, sort]);
  const total = useMemo(() => summarize(rows), [rows]);
  const scenarios = new Set(rows.map((r) => r.scenario)).size;

  const versions = useMemo(() => versioned(groups), [groups]);

  return (
    <TooltipLayer>
      <Versioned.Provider value={versions}>
        <header className="bar">
          <div className="title">
            <Nav on="stats" />
            <span className="dim">{all ? `${all.length} games · ${new Set(all.map((r) => r.run)).size} runs` : ""}</span>
          </div>
        </header>
        <main className="stats">
          <section className="hero">
            <h1>How well do models play Marvel Champions?</h1>
            <p className="dim">
              Every recorded game, scored as <code>mcbench report</code> scores it. Filter and group below; a game opens on the board.
            </p>
          </section>
          {error && <div className="panel section">Could not load the games: {error}</div>}
          {!all && !error && <div className="loading">Loading games…</div>}
          {all && (
            <>
              <FilterBar rows={all} f={f} set={(nf) => setView({ f: nf, by })} by={by} setBy={(nb) => setView({ f, by: nb })} dates={dates} />
              {rows.length === 0 ? (
                <div className="panel section empty-state">No game matches these filters.</div>
              ) : (
                <>
                  <Kpis total={total} runs={new Set(rows.map((r) => r.run)).size} rows={rows} />
                  <Section
                    title="Leaderboard"
                    sub="Mean score with its standard error, how often each wins and how its games ended."
                    delay={1}
                    extra={<SortPicker sort={sort} setSort={setSort} />}
                  >
                    <Leaderboard groups={groups} showScenario={scenarios > 1} />
                  </Section>
                  <div className="two">
                    <Section title="Score criteria" sub="Mean of each criterion the score weighs (win = win rate)." delay={2}>
                      <Criteria groups={groups} showScenario={scenarios > 1} />
                    </Section>
                    <Section title="Every game" sub="Each dot is a game at its score; click one to replay it." delay={3}>
                      <Strip groups={groups} showScenario={scenarios > 1} />
                    </Section>
                  </div>
                  <Section title="How they played" sub="The measures of mcbench analyze: why games were lost, how turns were used and ended." delay={4}>
                    <BehaviourTable groups={groups} showScenario={scenarios > 1} />
                  </Section>
                  <Section title="Runs over time" sub="Mean score of each run; click one to open it on the board." delay={5}>
                    <Timeline rows={rows} />
                  </Section>
                  <Section title="Games" sub="Every game the filters let through." delay={6}>
                    <GamesTable rows={rows} />
                  </Section>
                </>
              )}
            </>
          )}
        </main>
      </Versioned.Provider>
    </TooltipLayer>
  );
}
