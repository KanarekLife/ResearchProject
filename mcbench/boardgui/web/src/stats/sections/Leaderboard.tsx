import { useRef, useState } from "react";
import { boardLink, byNumber } from "../../runs";
import type { Group } from "../aggregate";
import { compact, fixed, pct } from "../format";
import { GameTip } from "../GameTip";
import { Meter, Stack } from "../marks";
import { useFlip } from "../motion";
import { GroupName } from "../names";
import { OUTCOMES, outcomeOf, outcomeParts } from "../outcomes";
import { useTip } from "../tooltip";

export type SortKey = "score" | "win" | "games" | "tokens";

export function SortPicker({ sort, setSort }: { sort: SortKey; setSort: (s: SortKey) => void }) {
  const keys: [SortKey, string][] = [
    ["score", "Score"],
    ["win", "Win rate"],
    ["games", "Games"],
    ["tokens", "Tokens"],
  ];
  return (
    <div className="seg" role="radiogroup" aria-label="Sort by">
      <span className="seg-label">Sort</span>
      {keys.map(([k, l]) => (
        <button key={k} role="radio" aria-checked={sort === k} className={sort === k ? "on" : ""} onClick={() => setSort(k)}>
          {l}
        </button>
      ))}
    </div>
  );
}

export function Leaderboard({ groups, showScenario }: { groups: Group[]; showScenario: boolean }) {
  const ref = useRef<HTMLDivElement>(null);
  const tip = useTip();
  const [open, setOpen] = useState<string | null>(null);
  useFlip(ref, groups.map((g) => g.key).join());
  return (
    <div className="board-table" ref={ref}>
      <div className="lb-row lb-head">
        <span>#</span>
        <span>Configuration</span>
        <span className="num">Games</span>
        <span>Score</span>
        <span>Win rate</span>
        <span>Outcomes</span>
        <span className="num">Rounds to win</span>
        <span className="num">Tokens / game</span>
      </div>
      {groups.map((g, i) => (
        <div key={g.key} data-flip={g.key} className={"lb-item" + (open === g.key ? " open" : "")}>
          <div
            className="lb-row"
            role="button"
            tabIndex={0}
            aria-expanded={open === g.key}
            onClick={() => setOpen(open === g.key ? null : g.key)}
            onKeyDown={(e) => (e.key === "Enter" || e.key === " ") && (e.preventDefault(), setOpen(open === g.key ? null : g.key))}
          >
            <span className="rank">{i + 1}</span>
            <GroupName g={g} showScenario={showScenario} />
            <span className="num">{g.games.length}</span>
            <span
              {...tip(() => (
                <>
                  <b>{fixed(3)(g.score.mean)}</b> mean score{" "}
                  <span className="dim">
                    ± {g.score.se.toFixed(3)} s.e. over {g.graded} graded
                  </span>
                </>
              ))}
            >
              <Meter value={g.score.mean} se={g.score.se} label={fixed(3)(g.score.mean)} />
            </span>
            <span
              {...tip(() => (
                <>
                  <b>{pct(g.criteria.win)}</b> won{" "}
                  <span className="dim">
                    · {g.status.won ?? 0} of {g.graded}
                  </span>
                </>
              ))}
            >
              <Meter value={g.criteria.win} color="var(--series-2)" label={pct(g.criteria.win)} />
            </span>
            <Stack parts={outcomeParts(g)} tip={tip} />
            <span className="num">{fixed(1)(g.roundsToWin)}</span>
            <span className="num">{compact(g.tokensIn / g.games.length)}</span>
          </div>
          {open === g.key && (
            <div className="lb-games">
              {[...g.games]
                .sort((a, b) => byNumber(a.run + a.game, b.run + b.game))
                .map((r) => (
                  <a key={r.run + r.game} className={"pill " + outcomeOf(r)} href={boardLink(r.run, r.game)} {...tip(() => <GameTip r={r} />)}>
                    seed {r.seed}
                    {r.sample ? `/s${r.sample}` : ""} · {fixed(2)(r.score)}
                  </a>
                ))}
            </div>
          )}
        </div>
      ))}
      <Legend />
    </div>
  );
}

function Legend() {
  return (
    <div className="legend">
      <span className="key">
        <i style={{ background: "var(--series-1)" }} /> score
      </span>
      <span className="key">
        <i style={{ background: "var(--series-2)" }} /> win rate
      </span>
      {OUTCOMES.map((o) => (
        <span className="key" key={o.key}>
          <i style={{ background: o.color }} /> {o.label}
        </span>
      ))}
      <span className="dim">· click a row to list its games</span>
    </div>
  );
}
