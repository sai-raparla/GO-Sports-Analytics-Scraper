import type { Player } from "../types";

interface Props {
  player: Player;
  year?: number;
}

// Baseball Reference uses placeholder abbreviations like "2TM"/"3TM" for the
// combined row of a season split across multiple teams.
const MULTI_TEAM_RE = /^\d+TM$/;

// teamForSeason returns the team(s) the player was on in the given season,
// derived from the per-season batting/pitching rows rather than the bio block
// (which only reflects the player's *current* team). Returns undefined when no
// matching season row exists so the caller can fall back to the bio team.
function teamForSeason(player: Player, year?: number): string | undefined {
  if (!year) return undefined;
  const season = String(year);
  const rows = [...(player.batting ?? []), ...(player.pitching ?? [])];
  const teams: string[] = [];
  for (const row of rows) {
    if (row.season !== season || !row.team) continue;
    if (!teams.includes(row.team)) teams.push(row.team);
  }
  if (teams.length === 0) return undefined;
  // When real team rows exist alongside the combined "2TM" placeholder, prefer
  // listing the actual teams.
  const named = teams.filter((t) => !MULTI_TEAM_RE.test(t));
  return (named.length > 0 ? named : teams).join(" / ");
}

export function PlayerHeader({ player, year }: Props) {
  const seasonTeam = teamForSeason(player, year) ?? player.team;
  const facts: { label: string; value?: string }[] = [
    { label: "Position", value: player.position },
    { label: "Bats", value: player.bats },
    { label: "Throws", value: player.throws },
    { label: "Height", value: player.height },
    { label: "Weight", value: player.weight },
    { label: year ? `Team (${year})` : "Team", value: seasonTeam },
  ];

  return (
    <header className="player-header">
      <div>
        <h2 className="player-name">{player.name}</h2>
        <div className="player-id">{player.id}</div>
      </div>
      <ul className="player-facts">
        {facts
          .filter((f) => f.value)
          .map((f) => (
            <li key={f.label}>
              <span className="fact-label">{f.label}</span>
              <span className="fact-value">{f.value}</span>
            </li>
          ))}
      </ul>
    </header>
  );
}
