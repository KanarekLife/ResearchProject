// A run id is <UTC timestamp>-<player>[__<instructions>], see cmd run.
export interface RunName {
  id: string;
  date: string; // YYYY-MM-DD, empty when the id has no timestamp
  time: string;
  player: string;
  instructions: string;
}

const RUN_ID = /^(\d{4})(\d{2})(\d{2})T(\d{2})(\d{2})(\d{2})Z-(.*?)(?:__(.*))?$/;

export function parseRun(id: string): RunName {
  const m = RUN_ID.exec(id);
  if (!m) return { id, date: "", time: "", player: id, instructions: "" };
  return { id, date: `${m[1]}-${m[2]}-${m[3]}`, time: `${m[4]}:${m[5]}`, player: m[7], instructions: m[8] ?? "" };
}

// byNumber sorts seed2 before seed10.
export const byNumber = (a: string, b: string) => a.localeCompare(b, undefined, { numeric: true });

export function boardLink(run: string, game: string): string {
  return "./?" + new URLSearchParams({ run, game });
}
