import type { Row } from "./games";
import { fixed } from "./format";
import { outcomeOf } from "./outcomes";

export function GameTip({ r }: { r: Row }) {
  return (
    <div className="gtip">
      <div>
        <b>{fixed(3)(r.score)}</b> <span className={"res " + outcomeOf(r)}>{r.status}</span>
      </div>
      <div className="dim">
        {r.player.replace(/^model:/, "")} · {r.instr}
      </div>
      <div className="dim">
        {r.scenario} · seed {r.seed} · {r.rounds} rounds · {r.date} {r.time}
      </div>
      {r.error && <div className="err">{r.error}</div>}
    </div>
  );
}
