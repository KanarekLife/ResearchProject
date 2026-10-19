import type { GroupBy } from "./aggregate";
import type { Filters } from "./filter";

export function fromAddress(): { f: Filters; by: GroupBy } {
  const q = new URLSearchParams(location.search);
  const by = (q.get("by") as GroupBy) || "config";
  return {
    f: {
      from: q.get("from") ?? "",
      to: q.get("to") ?? "",
      model: q.getAll("player"),
      instr: q.getAll("instr"),
      scenario: q.getAll("scenario"),
      status: q.getAll("result"),
    },
    by,
  };
}

export function toAddress(f: Filters, by: GroupBy) {
  const q = new URLSearchParams();
  if (f.from) q.set("from", f.from);
  if (f.to) q.set("to", f.to);
  f.model.forEach((v) => q.append("player", v));
  f.instr.forEach((v) => q.append("instr", v));
  f.scenario.forEach((v) => q.append("scenario", v));
  f.status.forEach((v) => q.append("result", v));
  if (by !== "config") q.set("by", by);
  const s = q.toString();
  history.replaceState(null, "", s ? "?" + s : location.pathname);
}
