// The JSON the Go server sends (boardgui.Snapshot and session.View).

export interface Snapshot {
  run: string;
  game: string; // the game's name; following a run, it changes to the next game
  info?: GameInfo;
  end?: Result;
  error?: string;
  total: number;
  // Changes when the server replaced earlier steps: fetch them all again.
  rebuilds: number;
  from: number;
  steps: Step[];
}

export interface GameInfo {
  scenario: string;
  seed: number;
  sample: number;
  player: string;
}

export interface Result {
  status: string;
  error?: string;
  rounds: number;
  score: number;
  criteria?: Record<string, number>;
}

// The board at one decision; the last step of a finished game has no decision.
export interface Step {
  board: View;
  status: string;
  decision?: Decision;
  choice?: Chosen;
}

export interface Decision {
  kind: string;
  prompt: string;
  options: Option[];
}

export interface Option {
  id: number;
  text: string;
  key: string;
}

export interface Chosen {
  key: string;
  text: string;
  reasoning?: string;
  // The model's reasoning_content for this decision; scripted players have none.
  thinking?: string;
  events: string[] | null;
}

export interface View {
  round: number;
  phase: string;
  hero: Hero;
  hand: Card[];
  in_play: Card[];
  deck_size: number;
  discard: string[];
  face_down_encounter_cards: number;
  villain: Enemy;
  next_villain_stages: Enemy[] | null;
  main_scheme: Scheme;
  side_schemes: Scheme[];
  minions: Enemy[];
  encounter_deck_size: number;
  encounter_discard: string[];
  acceleration_tokens: number;
  next_card_cost_reduction?: number;
}

export interface Hero {
  name: string;
  form: string; // hero or alter-ego
  hp: number;
  max_hp: number;
  atk?: number;
  thw?: number;
  def?: number;
  rec?: number;
  hand_size: number;
  exhausted: boolean;
  statuses?: string[];
  changed_form_this_turn: boolean;
  ability?: string;
  other_form: string;
  other_form_ability?: string;
  attachments?: Card[];
}

export interface Card {
  name: string;
  label?: string;
  type: string;
  cost?: number;
  resources?: string[];
  hp?: number;
  max_hp?: number;
  atk?: number;
  thw?: number;
  atk_consequential?: number;
  thw_consequential?: number;
  counters?: number;
  exhausted?: boolean;
  statuses?: string[];
  text?: string;
}

export interface Enemy {
  name: string;
  label?: string;
  stage?: number;
  hp: number;
  max_hp: number;
  atk: number;
  sch: number;
  keywords?: string[];
  statuses?: string[];
  text?: string;
  attachments?: Card[];
}

export interface Scheme {
  name: string;
  label?: string;
  threat: number;
  threshold?: number;
  threat_per_round?: number;
  icons?: string[];
  text?: string;
  attachments?: Card[];
}

// The runs under the results directory and the selection to open by default
// (boardgui.Runs). An empty game follows the run's current game.
export interface Runs {
  runs: { id: string; games: { name: string; state: string }[] }[];
  run: string;
  game: string;
}

// Card names to MarvelCDB codes, in code order.
export type Codes = Record<string, string[]>;
