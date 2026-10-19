import { ratio, type Group } from "../aggregate";
import { fixed, pct } from "../format";
import { Meter } from "../marks";
import { GroupName } from "../names";

const BEHAVIOUR: { label: string; title: string; value: (g: Group) => number; share?: boolean; digits?: number }[] = [
  { label: "lost: scheme", title: "Games lost because the main scheme completed", value: (g) => g.b.scheme, digits: 0 },
  { label: "lost: defeated", title: "Games lost with the hero at 0 HP", value: (g) => g.b.defeated, digits: 0 },
  { label: "basic / round", title: "Hero basic attacks and thwarts per round", value: (g) => ratio(g.b.basic, g.b.rounds), digits: 2 },
  { label: "attack", title: "Share of hero basic powers that attacked", value: (g) => ratio(g.b.attacks, g.b.basic), share: true },
  {
    label: "hero def",
    title: "Share of defences made by the hero (it then skips its next basic power)",
    value: (g) => ratio(g.b.hero_defends, g.b.defends),
    share: true,
  },
  { label: "turns", title: "Turns ended", value: (g) => g.b.ends, digits: 0 },
  { label: "AE ends", title: "Share of turns that ended in alter-ego", value: (g) => (g.b.recorded ? ratio(g.b.ae_ends, g.b.ends) : NaN), share: true },
  {
    label: "AE risky",
    title: "Share of turns that ended in alter-ego within 4 threat of the main scheme's target",
    value: (g) => (g.b.recorded ? ratio(g.b.ae_risky, g.b.ends) : NaN),
    share: true,
  },
  {
    label: "idle hero",
    title: "Share of turns that ended with a hero basic power unused",
    value: (g) => (g.b.recorded ? ratio(g.b.idle_hero, g.b.ends) : NaN),
    share: true,
  },
  {
    label: "idle ally",
    title: "Share of turns that ended with an ally attack or thwart unused",
    value: (g) => (g.b.recorded ? ratio(g.b.idle_ally, g.b.ends) : NaN),
    share: true,
  },
  {
    label: "card left",
    title: "Share of turns that ended with a card that could still be played",
    value: (g) => (g.b.recorded ? ratio(g.b.card_left, g.b.ends) : NaN),
    share: true,
  },
  { label: "acts / turn", title: "Actions before ending the turn", value: (g) => (g.b.recorded ? ratio(g.b.actions, g.b.ends) : NaN), digits: 2 },
];

export function BehaviourTable({ groups, showScenario }: { groups: Group[]; showScenario: boolean }) {
  return (
    <div className="scroll-x">
      <table className="beh">
        <thead>
          <tr>
            <th>Configuration</th>
            <th className="num">graded</th>
            {BEHAVIOUR.map((c) => (
              <th key={c.label} className="num" title={c.title}>
                <span className="hint">{c.label}</span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {groups.map((g) => (
            <tr key={g.key}>
              <td>
                <GroupName g={g} showScenario={showScenario} />
              </td>
              <td className="num">{g.b.graded}</td>
              {BEHAVIOUR.map((c) => {
                const v = c.value(g);
                return (
                  <td key={c.label} className="num" title={c.title}>
                    {c.share ? <Meter value={v} label={pct(v)} color="var(--series-3)" /> : fixed(c.digits ?? 2)(v)}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
