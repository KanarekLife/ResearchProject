import { useEffect, useLayoutEffect, useRef, useState, type RefObject } from "react";

const reduced = () => matchMedia("(prefers-reduced-motion: reduce)").matches;

// useMounted is false for the first frame, so marks can grow from zero.
export function useMounted(): boolean {
  const [on, setOn] = useState(false);
  useEffect(() => {
    const id = requestAnimationFrame(() => setOn(true));
    return () => cancelAnimationFrame(id);
  }, []);
  return on;
}

export function useCountUp(value: number, ms = 700): number {
  const [shown, setShown] = useState(0);
  const from = useRef(0);
  useEffect(() => {
    if (!Number.isFinite(value) || reduced()) {
      setShown(value);
      from.current = value;
      return;
    }
    const start = performance.now();
    const a = Number.isFinite(from.current) ? from.current : 0;
    let id = 0;
    const tick = (t: number) => {
      const p = Math.min((t - start) / ms, 1);
      const v = a + (value - a) * (1 - Math.pow(1 - p, 3));
      setShown(v);
      from.current = v;
      if (p < 1) id = requestAnimationFrame(tick);
    };
    id = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(id);
  }, [value, ms]);
  return shown;
}

// useFlip animates [data-flip] children from their previous position.
export function useFlip(ref: RefObject<HTMLElement | null>, order: string) {
  const last = useRef(new Map<string, number>());
  useLayoutEffect(() => {
    const els = [...(ref.current?.querySelectorAll<HTMLElement>("[data-flip]") ?? [])];
    const now = new Map(els.map((el) => [el.dataset.flip!, el.getBoundingClientRect().top]));
    if (!reduced()) {
      for (const el of els) {
        const before = last.current.get(el.dataset.flip!);
        const dy = before === undefined ? 0 : before - now.get(el.dataset.flip!)!;
        if (!dy) continue;
        el.animate([{ transform: `translateY(${dy}px)` }, { transform: "none" }], { duration: 450, easing: "cubic-bezier(.2,.7,.2,1)" });
      }
    }
    last.current = now;
  }, [ref, order]);
}
