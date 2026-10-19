import type { Group } from "../aggregate";
import { CRITERIA } from "../games";
import { fixed } from "../format";
import { useMounted } from "../motion";
import { GroupName } from "../names";
import { useTip } from "../tooltip";

export function Criteria({ groups, showScenario }: { groups: Group[]; showScenario: boolean }) {
  const tip = useTip();
  const on = useMounted();
  return (
    <div className="heat">
      <div className="heat-row heat-head">
        <span />
        {CRITERIA.map(([c, l]) => (
          <span key={c}>{l}</span>
        ))}
      </div>
      {groups.map((g) => (
        <div className="heat-row" key={g.key}>
          <GroupName g={g} showScenario={showScenario} />
          {CRITERIA.map(([c, l]) => {
            const v = g.criteria[c];
            return (
              <span
                key={c}
                className="cell"
                tabIndex={0}
                style={{ background: on && Number.isFinite(v) ? `color-mix(in oklab, var(--seq-hi) ${Math.round(v * 100)}%, var(--seq-lo))` : "var(--seq-lo)" }}
                {...tip(() => (
                  <>
                    <b>{fixed(2)(v)}</b> {l} <span className="dim">· mean over {g.graded} graded</span>
                  </>
                ))}
              >
                {fixed(2)(v)}
              </span>
            );
          })}
        </div>
      ))}
    </div>
  );
}
