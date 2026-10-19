import { useEffect, useState } from "react";
import type { Selection } from "./Picker";
import type { GameInfo, Result, Runs, Snapshot, Step } from "./types";

const POLL_MS = 1000;

export interface Replay {
  steps: Step[];
  game?: string;
  info?: GameInfo;
  end?: Result;
  error?: string;
}

// useReplay fetches the selected game's steps, polling while it is played. An
// exported site (`mcbench site`) has no server: its games are files that never change.
export function useReplay(sel: Selection | null, runs: Runs | null) {
  const [replay, setReplay] = useState<Replay>({ steps: [] });
  const [disconnected, setDisconnected] = useState(false);
  const loaded = runs !== null;
  const isStatic = !!runs?.static;
  const file = isStatic ? stepsFile(runs!, sel) : "";
  useEffect(() => {
    if (!sel || !loaded) return;
    if (isStatic && !file) {
      setReplay({ steps: [], error: "no such run or game on this site" });
      return;
    }
    let steps: Step[] = [];
    let game = ""; // the game the steps held are from
    let rebuilds = 0; // the server's replay version of the steps held
    let timer = 0;
    let stopped = false;
    const query = `api/steps?run=${encodeURIComponent(sel.run)}&game=${encodeURIComponent(sel.game)}&from=`;
    async function poll() {
      try {
        // Re-fetch the last step too: its choice arrives with the next one.
        const res = await fetch(file || query + Math.max(steps.length - 1, 0), { cache: "no-store" });
        if (!res.ok) {
          if (!stopped) setReplay({ steps: [], error: (await res.text()).trim() });
          return;
        }
        const s: Snapshot = await res.json();
        if (stopped) return;
        if (s.rebuilds !== rebuilds || s.game !== game) {
          rebuilds = s.rebuilds;
          game = s.game;
          steps = [];
          if (s.from > 0) return poll();
        }
        steps = [...steps.slice(0, s.from), ...s.steps];
        setReplay({ steps, game, info: s.info, end: s.end, error: s.error });
        setDisconnected(false);
        // A followed run goes on to its next game, so keep asking.
        if (file || (sel!.game && (s.end || s.error))) return;
      } catch (e) {
        console.error(e);
        setDisconnected(true);
      }
      if (!stopped) timer = setTimeout(poll, POLL_MS);
    }
    poll();
    return () => {
      stopped = true;
      clearTimeout(timer);
    };
  }, [sel, loaded, isStatic, file]);
  return { replay, setReplay, disconnected };
}

function stepsFile(runs: Runs, sel: Selection | null): string {
  const games = runs.runs.find((r) => r.id === sel?.run)?.games ?? [];
  const game = sel?.game ? games.find((g) => g.name === sel.game) : games[games.length - 1];
  return game?.steps ?? "";
}
