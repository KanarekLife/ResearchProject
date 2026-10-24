import type { Row } from "./games";
import { compact, fixed } from "./format";
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
      {(r.tokens_in > 0 || r.tokens_out > 0) && (
        <div className="dim">
          tokens {compact(r.tokens_in + r.tokens_out)} · {compact(r.tokens_in)} in · {compact(r.tokens_out)} out{r.tokens_cached ? ` · ${compact(r.tokens_cached)} cached` : ""}
        </div>
      )}
      {r.error && <div className="err">{r.error}</div>}
    </div>
  );
}
