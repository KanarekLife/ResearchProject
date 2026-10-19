import { boardLink } from "../../runs";
import type { Group } from "../aggregate";
import { GameTip } from "../GameTip";
import { useMounted } from "../motion";
import { GroupName } from "../names";
import { outcomeOf } from "../outcomes";
import { useTip } from "../tooltip";

// jitter is deterministic so dots keep their place across renders.
function jitter(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) h = Math.imul(h ^ s.charCodeAt(i), 16777619);
  return ((h >>> 0) % 1000) / 1000;
}

export function Strip({ groups, showScenario }: { groups: Group[]; showScenario: boolean }) {
  const tip = useTip();
  const on = useMounted();
  return (
    <div className="strip">
      {groups.map((g) => (
        <div className="strip-row" key={g.key}>
          <GroupName g={g} showScenario={showScenario} />
          <div className="strip-track">
            {Number.isFinite(g.score.mean) && <span className="strip-mean" style={{ left: `${g.score.mean * 100}%` }} />}
            {g.games.map((r, i) => (
              <a
                key={r.run + r.game}
                href={boardLink(r.run, r.game)}
                className={"dot " + outcomeOf(r)}
                style={{ left: `${(on ? r.score : 0) * 100}%`, top: `${15 + jitter(r.run + r.game) * 70}%`, transitionDelay: `${Math.min(i * 12, 400)}ms` }}
                {...tip(() => <GameTip r={r} />)}
                aria-label={`seed ${r.seed}, ${r.status}, score ${r.score.toFixed(3)}`}
              />
            ))}
          </div>
        </div>
      ))}
      <div className="strip-row axis">
        <span />
        <div className="strip-axis">
          {[0, 0.25, 0.5, 0.75, 1].map((t) => (
            <span key={t} style={{ left: `${t * 100}%` }}>
              {t}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
