import type { Player } from "../types";

interface Props {
  player: Player;
}

export function PlayerHeader({ player }: Props) {
  const facts: { label: string; value?: string }[] = [
    { label: "Position", value: player.position },
    { label: "Bats", value: player.bats },
    { label: "Throws", value: player.throws },
    { label: "Height", value: player.height },
    { label: "Weight", value: player.weight },
    { label: "Team", value: player.team },
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
