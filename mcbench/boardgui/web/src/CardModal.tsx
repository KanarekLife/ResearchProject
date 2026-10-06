import { useContext, useState } from "react";
import { chipsOf, CodesContext, codeOf, IMAGE_URL, Tokens, type Shown } from "./Board";

// CardModal shows one card full size: the MarvelCDB image (or its text when
// there is none) beside our name, text and current counters. A double-sided
// card opens on its current side and can be flipped.
export function CardModal({ s, onClose }: { s: Shown; onClose: () => void }) {
  const [flipped, setFlipped] = useState(false);
  const codes = useContext(CodesContext);
  const c = flipped && s.back ? s.back : s.c;
  const code = codeOf(codes, c.name, s.stage);
  const [broken, setBroken] = useState("");
  const counters = !flipped && (!!c.max_hp || c.threat !== undefined || !!c.counters);
  const chips = flipped ? [] : chipsOf(c);
  return (
    <div className="modal" onClick={onClose}>
      <div className="modal-box" role="dialog" aria-label={c.name} onClick={(e) => e.stopPropagation()}>
        <button className="close" title="Close (Esc)" onClick={onClose}>✕</button>
        {code && broken !== code ? (
          <img key={code} className="full" src={IMAGE_URL + code + ".png"} alt={c.name} onError={() => setBroken(code)} />
        ) : (
          <div className="full text-full">
            <b>{c.label || c.name}</b>
            <span className="dim">{c.type}</span>
            <span>{c.text || "No image or text for this card."}</span>
          </div>
        )}
        <div className="details">
          <h3>{c.label || c.name}</h3>
          <div className="dim">{[c.type, c.cost !== undefined && `cost ${c.cost}`, c.resources?.join(" ")].filter(Boolean).join(" · ")}</div>
          {c.text && <p>{c.text}</p>}
          {counters && <div className="counters"><Tokens c={c} /></div>}
          {!flipped && s.stats && <div>{s.stats}</div>}
          {chips.length > 0 && <div>{chips.join(", ")}</div>}
          {!flipped && !!c.attachments?.length && <div>Attached: {c.attachments.map((a) => a.label || a.name).join(", ")}</div>}
          {s.back && (
            <button onClick={() => setFlipped(!flipped)}>{flipped ? "Show current side" : `Show other side (${s.back.name})`}</button>
          )}
        </div>
      </div>
    </div>
  );
}
