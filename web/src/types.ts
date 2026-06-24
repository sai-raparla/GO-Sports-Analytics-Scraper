export interface SearchResult {
  id: string;
  name: string;
  years?: string;
  active: boolean;
}

export interface SeasonBatting {
  season: string;
  age?: string;
  team?: string;
  league?: string;
  war?: string;
  g?: string;
  pa?: string;
  ab?: string;
  r?: string;
  h?: string;
  hr?: string;
  rbi?: string;
  sb?: string;
  ba?: string;
  obp?: string;
  slg?: string;
  ops?: string;
}

export interface SeasonPitching {
  season: string;
  age?: string;
  team?: string;
  war?: string;
  w?: string;
  l?: string;
  era?: string;
  g?: string;
  gs?: string;
  sv?: string;
  ip?: string;
  so?: string;
  whip?: string;
}

export interface Player {
  id: string;
  name: string;
  position?: string;
  bats?: string;
  throws?: string;
  height?: string;
  weight?: string;
  born?: string;
  team?: string;
  source_url: string;
  batting?: SeasonBatting[];
  pitching?: SeasonPitching[];
}

export interface BattingSummary {
  games: number;
  pa: number;
  ab: number;
  r: number;
  h: number;
  "2b": number;
  "3b": number;
  hr: number;
  rbi: number;
  bb: number;
  so: number;
  sb: number;
  cs: number;
  hbp: number;
  sf: number;
  tb: number;
  avg: number;
  obp: number;
  slg: number;
  ops: number;
}

export interface PitchingSummary {
  games: number;
  outs: number;
  ip: string;
  h: number;
  r: number;
  er: number;
  bb: number;
  so: number;
  hr: number;
  era: number;
  whip: number;
  k9: number;
  bb9: number;
}

export interface GameLog {
  player_id: string;
  year: number;
  type: string;
  date: string;
  team?: string;
  opponent?: string;
  stats: Record<string, string>;
}

export interface RecentResponse {
  player_id: string;
  type: "batting" | "pitching";
  year: number;
  days: number;
  from: string;
  to: string;
  games: number;
  batting?: BattingSummary;
  pitching?: PitchingSummary;
  game_logs: GameLog[];
}

export type StatType = "batting" | "pitching";
