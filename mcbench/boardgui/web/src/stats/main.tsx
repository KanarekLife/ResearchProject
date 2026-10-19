import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Stats } from "./Stats";
import "../app.css";
import "./stats.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Stats />
  </StrictMode>,
);
