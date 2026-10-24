import { useEffect, useRef, useState } from "react";
import type { Chosen, Result, Step } from "./types";

// thousands and secs format token usage and trace times; the board only
// formats them here.
const thousands = (x: number) => Math.round(x).toLocaleString("en-US");
const secs = (ms?: number) => (!ms || ms <= 0 ? undefined : ms < 10000 ? `${(ms / 1000).toFixed(1)}s` : `${Math.round(ms / 1000)}s`);

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
  // An old trace records "play:X|pay:Y" where the step offers "play:X".
  const isChosen = (key: string) => !!s.choice && (s.choice.key === key || s.choice.key.startsWith(key + "|pay:"));
  return (
    <>
      <h3>{s.decision ? `Decision ${i + 1} · ${s.decision.kind}` : "Game over"}</h3>
      <div className="dim">{`round ${s.board.round} · ${s.board.phase} phase`}</div>
      {s.decision && (
        <>
          <div className="prompt">{s.decision.prompt}</div>
          <ol className="options">
            {s.decision.options.map((o) => (
              <li key={o.id} className={isChosen(o.key) ? "chosen" : undefined} title={o.key}>
                {(isChosen(o.key) ? "✓ " : "") + o.text}
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
      {(s.choice?.tokens_in != null || s.choice?.tokens_out != null || s.choice?.tokens_cached != null) && (
        <div className="tok">
          in {thousands(s.choice.tokens_in ?? 0)} · out {thousands(s.choice.tokens_out ?? 0)}
          {s.choice.tokens_cached ? ` · cached ${thousands(s.choice.tokens_cached)}` : ""}
        </div>
      )}
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
  const u = e.usage;
  const out = u?.output_tokens ?? 0;
  if (u) {
    if (u.requests) rows.push(["requests", u.requests]);
    if (u.input_tokens) rows.push(["tokens in", thousands(u.input_tokens)]);
    if (out) rows.push(["tokens out", thousands(out)]);
    if (u.cached_tokens) rows.push(["cached", thousands(u.cached_tokens)]);
  }
  if (e.ms && out) rows.push(["tok/s", Math.round(out / (e.ms / 1000))]);
  if (e.ms) rows.push(["time", secs(e.ms) ?? ""]);
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
