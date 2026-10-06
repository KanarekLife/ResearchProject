import { createContext, useContext, useState, type ReactNode } from "react";
import type { Card, Codes, Enemy, Scheme, View } from "./types";

export const IMAGE_URL = "https://marvelcdb.com/bundles/cards/";

export const CodesContext = createContext<Codes>({});

// Any card the board draws: a player card, an enemy, a scheme or the hero.
export type AnyCard = { name: string } & Partial<Card & Enemy & Scheme>;

// Shown is a card opened full size: as drawn on the board, plus the other side
// of a double-sided card.
export interface Shown {
  c: AnyCard;
  stage?: number;
  stats?: string;
  back?: AnyCard;
}

// OpenContext opens a card full size.
export const OpenContext = createContext<(s: Shown) => void>(() => {});

interface CardOpts {
  size?: "big";
  wide?: boolean;
  hand?: boolean;
  stage?: number;
  stats?: string;
  back?: AnyCard;
}

// Board draws the zones on a fixed grid: every zone keeps its place and size
// from step to step, and cards that do not fit scroll inside their zone.
export function Board({ v }: { v: View }) {
  const vil = v.villain;
  const ms = v.main_scheme;
  const hero = v.hero;
  const isHero = hero.form === "hero";
  const heroStats = isHero ? stat([["ATK", hero.atk], ["THW", hero.thw], ["DEF", hero.def]]) : stat([["REC", hero.rec]]);
  const heroBack = { name: hero.other_form, type: isHero ? "alter-ego" : "hero", text: hero.other_form_ability };
  const next = (v.next_villain_stages ?? []).map((s) => `${roman(s.stage)} (${s.max_hp} HP)`).join(", ");
  return (
    <section className="board">
      <Zone area="villain" side="encounter" title={next ? `Villain · next ${next}` : "Villain"}>
        <CardView c={vil} size="big" stage={vil.stage} stats={`stage ${roman(vil.stage)} · ` + stat([["ATK", vil.atk], ["SCH", vil.sch]])} />
      </Zone>
      <Zone area="scheme" side="encounter" title="Main scheme">
        <CardView c={ms} wide size="big" stats={ms.threat_per_round ? `+${ms.threat_per_round} threat per round` : ""} />
      </Zone>
      <Zone area="sides" side="encounter" title="Side schemes">
        {v.side_schemes.map((s, i) => <CardView key={i} c={s} wide />)}
      </Zone>
      <Zone area="encounter" side="encounter" title="Encounter" piles>
        <Pile label="deck" value={v.encounter_deck_size} />
        <Pile label="discard" value={v.encounter_discard.length} list={v.encounter_discard} />
        <Pile label="face-down dealt" value={v.face_down_encounter_cards} />
        <Pile label="acceleration" value={v.acceleration_tokens} />
      </Zone>
      <Zone area="minions" title="Minions engaged">
        {v.minions.map((m, i) => <CardView key={i} c={m} stats={stat([["ATK", m.atk], ["SCH", m.sch]])} />)}
      </Zone>
      <Zone area="hero" side="player" title={isHero ? "Hero" : "Alter-ego"}>
        <CardView c={{ ...hero, type: hero.form, text: hero.ability }} size="big" stats={heroStats} back={heroBack} />
      </Zone>
      <Zone area="allies" side="player" title="Allies">
        {v.in_play.filter((c) => c.type === "ally").map((c, i) => <CardView key={i} c={c} stats={allyStats(c)} />)}
      </Zone>
      <Zone area="upgrades" side="player" title="Upgrades & supports">
        {v.in_play.filter((c) => c.type !== "ally").map((c, i) => <CardView key={i} c={c} />)}
      </Zone>
      <Zone area="player" side="player" title="Player" piles>
        <Pile label="deck" value={v.deck_size} />
        <Pile label="discard" value={v.discard.length} list={v.discard} />
        <Pile label="hand size" value={hero.hand_size} />
        {!!v.next_card_cost_reduction && <Pile label="next card cost" value={"−" + v.next_card_cost_reduction} />}
        {hero.changed_form_this_turn && <Pile label="changed form" value="yes" />}
      </Zone>
      <Zone area="hand" title={`Hand (${v.hand.length})`}>
        {v.hand.map((c, i) => <CardView key={i} c={c} hand stats={allyStats(c)} />)}
      </Zone>
    </section>
  );
}

function Zone(p: { area: string; side?: string; title: string; piles?: boolean; children: ReactNode }) {
  const empty = Array.isArray(p.children) && p.children.length === 0;
  return (
    <div className={["zone", p.side].filter(Boolean).join(" ")} style={{ gridArea: p.area }}>
      <h2 title={p.title}>{p.title}</h2>
      <div className={p.piles ? "piles" : "cards"}>{empty ? <span className="empty">none</span> : p.children}</div>
    </div>
  );
}

function Pile({ label, value, list }: { label: string; value: ReactNode; list?: string[] }) {
  return (
    <div className="pile" title={list?.length ? list.join("\n") : undefined}>
      <span>{label}</span>
      <b>{value}</b>
    </div>
  );
}

// CardView draws one card with its tokens (HP, threat, counters, cost), status
// chips and, over its foot, the names of anything attached to it. Clicking opens it full size.
function CardView({ c, ...opts }: { c: AnyCard } & CardOpts) {
  const open = useContext(OpenContext);
  const classes = ["card", opts.size, opts.wide && "wide", c.exhausted && "exhausted"].filter(Boolean).join(" ");
  return (
    <div className={classes} title={[c.label || c.name, c.type].filter(Boolean).join(" · ")}>
      <div className="face-wrap" onClick={() => open({ c, stage: opts.stage, stats: opts.stats, back: opts.back })}>
        <Face c={c} stage={opts.stage} />
        <Tokens c={c} cost={opts.hand} />
        <div className="chips">
          {chipsOf(c).map((s, i) => <span key={i} className={"chip " + s}>{s}</span>)}
        </div>
        {!!c.attachments?.length && (
          <div className="attachments">
            {c.attachments.map((a, i) => (
              <button key={i} title="Attached; click to open" onClick={(e) => {
                  e.stopPropagation();
                  open({ c: a });
                }}>
                + {a.label || a.name}
              </button>
            ))}
          </div>
        )}
      </div>
      <div className="name">{c.label || c.name}</div>
      {opts.stats && <div className="stats">{opts.stats}</div>}
    </div>
  );
}

export function Tokens({ c, cost }: { c: AnyCard; cost?: boolean }) {
  return (
    <div className="tokens">
      {!!c.max_hp && <span className="token hp" title="hit points">{`${c.hp}/${c.max_hp}`}</span>}
      {c.threat !== undefined && <span className="token threat" title="threat">{c.threshold ? `${c.threat}/${c.threshold}` : c.threat}</span>}
      {!!c.counters && <span className="token counter" title="counters">{c.counters}</span>}
      {cost && c.cost !== undefined && <span className="token cost" title="cost">{c.cost}</span>}
    </div>
  );
}

export function chipsOf(c: AnyCard): string[] {
  const chips = [...(c.keywords ?? []), ...(c.icons ?? []), ...(c.statuses ?? [])];
  if (c.exhausted) chips.push("exhausted");
  return chips;
}

// Face is the card's MarvelCDB image, or its text when it has no code or image.
export function Face({ c, stage }: { c: AnyCard; stage?: number }) {
  const code = codeOf(useContext(CodesContext), c.name, stage);
  const [broken, setBroken] = useState("");
  if (!code || broken === code) {
    return (
      <div className="face">
        <div className="text-card">
          <b>{c.name}</b>
          <span className="type">{c.type || ""}</span>
          {c.cost !== undefined && <span>cost {c.cost}</span>}
          {c.text && <span>{c.text}</span>}
        </div>
      </div>
    );
  }
  return (
    <div className="face">
      <img src={IMAGE_URL + code + ".png"} alt={c.name} loading="lazy" onError={() => setBroken(code)} />
    </div>
  );
}

// codeOf picks the card's code; only villain stages share a name.
export function codeOf(codes: Codes, name: string, stage?: number): string {
  const cs = codes[name];
  if (!cs?.length) return "";
  return cs[Math.min((stage || 1) - 1, cs.length - 1)];
}

function stat(parts: [string, number | undefined][]): string {
  return parts.filter(([, v]) => v !== undefined && v !== null).map(([k, v]) => `${k} ${v}`).join(" · ");
}

function allyStats(c: Card): string {
  return c.type === "ally" ? stat([["ATK", c.atk], ["THW", c.thw]]) : "";
}

function roman(n = 0): string {
  return ["", "I", "II", "III", "IV", "V"][n] || String(n);
}
