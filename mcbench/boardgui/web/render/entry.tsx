// POC: static render of the board GUI's Board component for vision input.
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { Board, CodesContext } from "../src/Board";
import type { Codes, View } from "../src/types";

export function renderBoard(v: View, codes: Codes): string {
  return renderToStaticMarkup(
    createElement(CodesContext.Provider, { value: codes }, createElement("div", { className: "layout" }, createElement(Board, { v }))),
  );
}
