import { createContext, useContext } from "react";
import { parseRun } from "../runs";
import type { Group } from "./aggregate";

// Instruction sets seen in more than one version, so their groups name it.
export const Versioned = createContext<Set<string>>(new Set());

export function versioned(groups: Group[]): Set<string> {
  const seen = new Map<string, Set<string>>();
  for (const g of groups) if (g.version) seen.set(g.instr, (seen.get(g.instr) ?? new Set()).add(g.version));
  return new Set([...seen].filter(([, v]) => v.size > 1).map(([k]) => k));
}

function PlayerName({ player }: { player: string }) {
  const [name, ...settings] = player.replace(/^model:/, "").split("@");
  return (
    <span className="player" title={player}>
      {name}
      {settings.length > 0 && <span className="settings"> {settings.join(" · ")}</span>}
    </span>
  );
}

export function GroupName({ g, showScenario }: { g: Group; showScenario: boolean }) {
  const r = g.run ? parseRun(g.run) : null;
  const showVersion = useContext(Versioned).has(g.instr);
  return (
    <div className="gname">
      {g.player && <PlayerName player={g.player} />}
      <span className="tags">
        {r && (
          <span className="tag when">
            {r.date} {r.time}
          </span>
        )}
        {g.instr &&
          g.instr.split(",").map((t) => (
            <span className="tag" key={t}>
              {t}
            </span>
          ))}
        {g.version && showVersion && (
          <span className="tag ver" title={"instructions version " + g.version}>
            v·{g.version.slice(0, 6)}
          </span>
        )}
        {showScenario && <span className="tag scen">{g.scenario}</span>}
      </span>
    </div>
  );
}
