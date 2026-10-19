export function Nav({ on }: { on: "board" | "stats" }) {
  return (
    <nav className="pages" aria-label="Pages">
      <a href="./" className={on === "board" ? "on" : undefined}>Board</a>
      <a href="stats.html" className={on === "stats" ? "on" : undefined}>Statistics</a>
    </nav>
  );
}
