import { createContext, useContext, useState, type ReactNode } from "react";

interface Tip {
  x: number;
  y: number;
  body: ReactNode;
}

const TipContext = createContext<(t: Tip | null) => void>(() => {});

export function TooltipLayer({ children }: { children: ReactNode }) {
  const [tip, setTip] = useState<Tip | null>(null);
  const left = tip ? Math.min(tip.x + 14, innerWidth - 300) : 0;
  return (
    <TipContext.Provider value={setTip}>
      {children}
      <div className={"tip" + (tip ? " on" : "")} style={{ left, top: tip ? tip.y + 14 : 0 }} role="tooltip">
        {tip?.body}
      </div>
    </TipContext.Provider>
  );
}

export function useTip() {
  const set = useContext(TipContext);
  return (body: () => ReactNode) => ({
    onPointerMove: (e: React.PointerEvent) => set({ x: e.clientX, y: e.clientY, body: body() }),
    onPointerLeave: () => set(null),
    onFocus: (e: React.FocusEvent) => {
      const r = (e.target as HTMLElement).getBoundingClientRect();
      set({ x: r.left, y: r.bottom, body: body() });
    },
    onBlur: () => set(null),
  });
}
