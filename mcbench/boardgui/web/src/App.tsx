// The board for one step of a game trace. The server replays the trace; this
// page only draws the steps it is given and asks for new ones while the game
// is still being played. The game shown is chosen in the address
// (?run=…&game=…, no game: follow the run) and the step by #N.
import { useCallback, useEffect, useRef, useState } from "react";
import { Board, CodesContext, OpenContext, type Shown } from "./Board";
import { CardModal } from "./CardModal";
import { Nav } from "./Nav";
import { Picker, type Selection } from "./Picker";
import { Side } from "./Side";
import type { Codes, Runs } from "./types";
import { useReplay } from "./useReplay";

const RUNS_POLL_MS = 5000;

// fromAddress reads the selection in the address, if any.
function fromAddress(): Selection | null {
  const q = new URLSearchParams(location.search);
  const run = q.get("run");
  return run ? { run, game: q.get("game") ?? "" } : null;
}

export function App() {
  const [codes, setCodes] = useState<Codes>({});
  const [runs, setRuns] = useState<Runs | null>(null);
  const [sel, setSel] = useState<Selection | null>(fromAddress);
  const [cur, setCur] = useState(0);
  const [shown, setShown] = useState<Shown | null>(null);
  // A #N in the address opens step N (1-based) instead of following the game.
  const wanted = useRef(parseInt(location.hash.slice(1), 10) - 1);
  const [follow, setFollow] = useState(!(wanted.current >= 0));
  const shownGame = useRef<string | undefined>(undefined);
  const { replay, setReplay, disconnected } = useReplay(sel, runs);

  useEffect(() => {
    fetch("api/cards").then((r) => r.json()).then(setCodes).catch(() => {});
  }, []);

  // The picker's list, refreshed so new runs and games show up. Without a
  // selection in the address the page opens the command line's.
  useEffect(() => {
    const load = () =>
      fetch("api/runs", { cache: "no-store" })
        .then((r) => r.json())
        .then((r: Runs) => {
          setRuns(r);
          setSel((s) => s ?? { run: r.run, game: r.game });
          if (r.static) clearInterval(timer);
        })
        .catch(() => {});
    const timer = setInterval(load, RUNS_POLL_MS);
    load();
    const onPop = () => {
      wanted.current = parseInt(location.hash.slice(1), 10) - 1;
      setReplay({ steps: [] });
      setSel(fromAddress());
      if (!fromAddress()) load();
    };
    addEventListener("popstate", onPop);
    return () => {
      clearInterval(timer);
      removeEventListener("popstate", onPop);
    };
  }, []);


  const { steps, end, error } = replay;
  const n = steps.length;
  const run = runs?.runs.find((r) => r.id === sel?.run);
  const local = !!run?.local;

  // New steps move to the last one when following, or to the #N asked for. A
  // new game opens on its first step when not following.
  useEffect(() => {
    const newGame = replay.game !== shownGame.current;
    shownGame.current = replay.game;
    if (wanted.current >= 0 && wanted.current < n) {
      setCur(wanted.current);
      wanted.current = -1;
    } else if (follow && local) {
      setCur(Math.max(n - 1, 0));
    } else if (newGame && wanted.current < 0) {
      setCur(0);
    }
  }, [replay, follow, local, n]);

  const go = useCallback(
    (i: number, keepFollow = false) => {
      const to = Math.min(Math.max(i, 0), Math.max(n - 1, 0));
      setCur(to);
      if (!keepFollow) setFollow(to === n - 1);
      history.replaceState(null, "", "#" + (to + 1));
    },
    [n],
  );

  const toggleFollow = useCallback(
    (on: boolean) => {
      setFollow(on);
      if (on) go(n - 1, true);
    },
    [go, n],
  );

  // choose opens another game; the back button returns to the one before.
  const choose = (s: Selection) => {
    const q = new URLSearchParams({ run: s.run });
    if (s.game) q.set("game", s.game);
    history.pushState(null, "", "?" + q);
    wanted.current = -1;
    setReplay({ steps: [] });
    setSel(s);
  };

  useEffect(() => {
    const keys: Record<string, () => void> = {
      ArrowLeft: () => go(cur - 1),
      ArrowRight: () => go(cur + 1),
      Home: () => go(0),
      End: () => go(n - 1),
    };
    if (local) keys.f = () => toggleFollow(!follow);
    const onKey = (e: KeyboardEvent) => {
      // While a card is open, Esc closes it and no other key steps the game.
      if (shown) {
        if (e.key === "Escape") setShown(null);
        return;
      }
      if (!keys[e.key] || (e.target as HTMLElement).tagName === "SELECT") return;
      e.preventDefault();
      keys[e.key]();
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [go, toggleFollow, cur, n, follow, local, shown]);

  const state = end ? end.status : error ? "stopped" : "live";
  const step = steps[cur];
  return (
    <CodesContext.Provider value={codes}>
      <OpenContext.Provider value={setShown}>
        <header className="bar">
          <div className="title">
            <Nav on="board" />
            <Picker runs={runs} sel={sel} following={replay.game} choose={choose} />
            <span className={"badge " + state}>{disconnected ? "disconnected" : state}</span>
            {run?.version && <span className="badge" title="Commit the run was played on">{run.version}</span>}
          </div>
          <nav className="nav">
            <button title="First step (Home)" disabled={cur <= 0} onClick={() => go(0)}>⏮</button>
            <button title="Previous step (←)" disabled={cur <= 0} onClick={() => go(cur - 1)}>◀</button>
            <input id="slider" type="range" min={0} max={Math.max(n - 1, 0)} value={cur} aria-label="Step" onChange={(e) => go(Number(e.target.value))} />
            <button title="Next step (→)" disabled={cur >= n - 1} onClick={() => go(cur + 1)}>▶</button>
            <button title="Last step (End)" disabled={cur >= n - 1} onClick={() => go(n - 1)}>⏭</button>
            <span id="counter" className="dim">{n ? `${cur + 1} / ${n}` : "0 / 0"}</span>
            {local && (
              <label title="Jump to each new step as the game is played (F)" className={end && sel?.game ? "gone" : undefined}>
                <input type="checkbox" checked={follow} onChange={(e) => toggleFollow(e.target.checked)} /> follow
              </label>
            )}
          </nav>
        </header>
        <main className="layout">
          {step && <Board v={step.board} />}
          {step && <Side steps={steps} cur={cur} go={go} end={end} />}
        </main>
        {error && <div className="error">Replay stopped: {error}</div>}
        {shown && <CardModal s={shown} onClose={() => setShown(null)} />}
      </OpenContext.Provider>
    </CodesContext.Provider>
  );
}
