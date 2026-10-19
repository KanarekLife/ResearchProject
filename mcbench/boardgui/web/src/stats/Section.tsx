import type { ReactNode } from "react";

export function Section({ title, sub, delay, extra, children }: { title: string; sub: string; delay: number; extra?: ReactNode; children: ReactNode }) {
  return (
    <section className="panel section" style={{ animationDelay: `${delay * 70}ms` }}>
      <div className="section-head">
        <div>
          <h2>{title}</h2>
          <p className="dim">{sub}</p>
        </div>
        {extra}
      </div>
      {children}
    </section>
  );
}
