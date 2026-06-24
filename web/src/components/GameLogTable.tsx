import type { GameLog, StatType } from "../types";

interface Props {
  logs: GameLog[];
  type: StatType;
}

interface Col {
  key: string;
  label: string;
}

const battingCols: Col[] = [
  { key: "b_ab", label: "AB" },
  { key: "b_r", label: "R" },
  { key: "b_h", label: "H" },
  { key: "b_doubles", label: "2B" },
  { key: "b_triples", label: "3B" },
  { key: "b_hr", label: "HR" },
  { key: "b_rbi", label: "RBI" },
  { key: "b_bb", label: "BB" },
  { key: "b_so", label: "SO" },
  { key: "b_sb", label: "SB" },
];

const pitchingCols: Col[] = [
  { key: "p_game_decision", label: "Dec" },
  { key: "p_ip", label: "IP" },
  { key: "p_h", label: "H" },
  { key: "p_r", label: "R" },
  { key: "p_er", label: "ER" },
  { key: "p_bb", label: "BB" },
  { key: "p_so", label: "SO" },
  { key: "p_hr", label: "HR" },
];

export function GameLogTable({ logs, type }: Props) {
  const cols = type === "pitching" ? pitchingCols : battingCols;

  if (logs.length === 0) {
    return (
      <div className="empty">
        No games in this window. Try widening the date range or switching the
        season.
      </div>
    );
  }

  return (
    <div className="table-wrap">
      <table className="gamelog">
        <thead>
          <tr>
            <th className="sticky-col">Date</th>
            <th>Opp</th>
            {cols.map((c) => (
              <th key={c.key}>{c.label}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {logs.map((l, i) => (
            <tr key={`${l.date}-${i}`}>
              <td className="sticky-col">{l.date}</td>
              <td>{l.opponent || "-"}</td>
              {cols.map((c) => (
                <td key={c.key}>{l.stats[c.key] ?? "-"}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
