import { useEffect, useRef, useState } from "react";
import type { Chosen, Result, Step } from "./types";

// Side is the decision panel, a fixed box whose content scrolls. One tab shows
// the current decision: the prompt, the options with the chosen one marked,
// the player's reasoning, what followed and, at the end, the result. The other
// lists every decision of the current round with its choice and reasoning.
export function Side({ steps, cur, go, end }: { steps: Step[]; cur: number; go: (i: number) => void; end?: Result }) {
  const [tab, setTab] = useState<"decision" | "round">("decision");
  const body = useRef<HTMLDivElement>(null);
  const s = steps[cur];
  useEffect(() => {
    body.current?.querySelector(".chosen, .current")?.scrollIntoView({ block: "nearest" });
  }, [cur, tab, s.choice?.key]);

  const round = s.board.round;
  return (
    <aside className="side">
      <div className="tabs">
        <button className={tab === "decision" ? "on" : ""} onClick={() => setTab("decision")}>Decision</button>
        <button className={tab === "round" ? "on" : ""} onClick={() => setTab("round")}>Round {round} reasoning</button>
      </div>
      <div className="side-body" ref={body}>
        {tab === "decision" ? (
          <DecisionView s={s} i={cur} last={cur === steps.length - 1} end={end} />
        ) : (
          <ol className="round-log">
            {steps.map((t, i) =>
              t.board.round !== round ? null : (
                <li key={i} className={i === cur ? "current" : undefined} onClick={() => go(i)}>
                  <div className="dim">{t.decision ? `#${i + 1} · ${t.decision.kind} · ${t.board.phase} phase` : `#${i + 1} · game over`}</div>
                  {t.decision && <div className="picked">{t.choice ? t.choice.text : end ? "No choice recorded." : "Waiting for the player…"}</div>}
                  {t.choice && <Reasoning c={t.choice} />}
                </li>
              ),
            )}
          </ol>
        )}
      </div>
    </aside>
  );
}

function DecisionView({ s, i, last, end }: { s: Step; i: number; last: boolean; end?: Result }) {
  const chosen = s.choice?.key;
  return (
    <>
      <h3>{s.decision ? `Decision ${i + 1} · ${s.decision.kind}` : "Game over"}</h3>
      <div className="dim">{`round ${s.board.round} · ${s.board.phase} phase`}</div>
      {s.decision && (
        <>
          <div className="prompt">{s.decision.prompt}</div>
          <ol className="options">
            {s.decision.options.map((o) => (
              <li key={o.id} className={o.key === chosen ? "chosen" : undefined} title={o.key}>
                {(o.key === chosen ? "✓ " : "") + o.text}
              </li>
            ))}
          </ol>
          {!s.choice && <p className="waiting">{end ? "No choice recorded." : "Waiting for the player…"}</p>}
        </>
      )}
      {s.choice && (
        <>
          <h2>Reasoning</h2>
          <Reasoning c={s.choice} />
        </>
      )}
      {!!s.choice?.events?.length && (
        <>
          <h2>Then</h2>
          <ul className="events">
            {s.choice.events.map((e, k) => <li key={k}>{e}</li>)}
          </ul>
        </>
      )}
      {end && (!s.decision || last) && <ResultTable e={end} />}
    </>
  );
}

// Reasoning is what the player said about its choice and, for a model, its
// thinking before answering, folded.
function Reasoning({ c }: { c: Chosen }) {
  return (
    <>
      {c.reasoning ? <div className="reason">{c.reasoning}</div> : <div className="dim no-reason">No reasoning recorded.</div>}
      {c.thinking && (
        <details className="thinking" onClick={(e) => e.stopPropagation()}>
          <summary>Model thinking</summary>
          <div>{c.thinking}</div>
        </details>
      )}
    </>
  );
}

function ResultTable({ e }: { e: Result }) {
  const rows: [string, string | number][] = [["status", e.status], ["rounds", e.rounds], ["score", e.score.toFixed(3)]];
  for (const [k, v] of Object.entries(e.criteria ?? {})) rows.push([k, v.toFixed(3)]);
  if (e.error) rows.push(["error", e.error]);
  return (
    <div className="result">
      <h2>Result</h2>
      <table>
        <tbody>
          {rows.map(([k, v]) => (
            <tr key={k}>
              <td className="dim">{k}</td>
              <td>{v}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
