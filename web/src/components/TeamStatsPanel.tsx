import type { StatType, Team } from "../types";

interface Props {
  team: Team;
  type: StatType;
  onTypeChange: (type: StatType) => void;
}

const BATTING_CARDS = [
  ["b_r", "R"],
  ["b_hr", "HR"],
  ["b_rbi", "RBI"],
  ["b_h", "H"],
  ["b_batting_avg", "AVG"],
  ["b_onbase_perc", "OBP"],
  ["b_slugging_perc", "SLG"],
  ["b_onbase_plus_slugging", "OPS"],
  ["b_bb", "BB"],
  ["b_so", "SO"],
  ["b_sb", "SB"],
  ["b_war", "WAR"],
];

const PITCHING_CARDS = [
  ["p_earned_run_avg", "ERA"],
  ["p_whip", "WHIP"],
  ["p_w", "W"],
  ["p_l", "L"],
  ["p_sv", "SV"],
  ["p_ip", "IP"],
  ["p_so", "SO"],
  ["p_bb", "BB"],
  ["p_hr", "HR"],
  ["p_fip", "FIP"],
  ["p_war", "WAR"],
  ["p_earned_run_avg_plus", "ERA+"],
];

const LABELS: Record<string, string> = {
  b_ab: "AB",
  b_batting_avg: "AVG",
  b_bb: "BB",
  b_cs: "CS",
  b_doubles: "2B",
  b_games: "G",
  b_gidp: "GIDP",
  b_h: "H",
  b_hbp: "HBP",
  b_hr: "HR",
  b_ibb: "IBB",
  b_onbase_perc: "OBP",
  b_onbase_plus_slugging: "OPS",
  b_onbase_plus_slugging_plus: "OPS+",
  b_pa: "PA",
  b_r: "R",
  b_rbi: "RBI",
  b_sb: "SB",
  b_slugging_perc: "SLG",
  b_so: "SO",
  b_tb: "TB",
  b_triples: "3B",
  b_war: "WAR",
  p_bb: "BB",
  p_bb_per_nine: "BB/9",
  p_bfp: "BF",
  p_earned_run_avg: "ERA",
  p_earned_run_avg_plus: "ERA+",
  p_er: "ER",
  p_fip: "FIP",
  p_g: "G",
  p_gs: "GS",
  p_h: "H",
  p_hr: "HR",
  p_hr_per_nine: "HR/9",
  p_ip: "IP",
  p_l: "L",
  p_r: "R",
  p_so: "SO",
  p_so_per_nine: "K/9",
  p_sv: "SV",
  p_w: "W",
  p_war: "WAR",
  p_whip: "WHIP",
  p_win_loss_perc: "W-L%",
};

export function TeamStatsPanel({ team, type, onTypeChange }: Props) {
  const stats = type === "batting" ? team.batting_totals : team.pitching_totals;
  const cardDefs = type === "batting" ? BATTING_CARDS : PITCHING_CARDS;
  const cards = cardDefs
    .map(([key, label]) => ({ key, label, value: stats?.[key] }))
    .filter((card) => card.value);

  const rows = Object.entries(stats ?? {}).sort(([a], [b]) =>
    statLabel(a).localeCompare(statLabel(b)),
  );

  return (
    <main className="content">
      <section className="team-header">
        <div>
          <h2 className="player-name">{team.name}</h2>
          <span className="player-id">
            {team.id} · {team.year}
          </span>
        </div>
        <div className="toggle">
          {(["batting", "pitching"] as StatType[]).map((t) => (
            <button
              key={t}
              className={`toggle-btn ${type === t ? "active" : ""}`}
              onClick={() => onTypeChange(t)}
            >
              {t[0].toUpperCase() + t.slice(1)}
            </button>
          ))}
        </div>
      </section>

      <section className="summary">
        <div className="summary-heading">
          <h3>{type === "batting" ? "Batting totals" : "Pitching totals"}</h3>
        </div>
        <div className="stat-grid">
          {cards.map((card) => (
            <div className="stat-card" key={card.key}>
              <div className="stat-value">{card.value}</div>
              <div className="stat-label">{card.label}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="games-section">
        <h3>All totals</h3>
        <div className="table-wrap">
          <table className="gamelog stats-table">
            <thead>
              <tr>
                <th>Stat</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              {rows.map(([key, value]) => (
                <tr key={key}>
                  <td>{statLabel(key)}</td>
                  <td>{value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </main>
  );
}

function statLabel(key: string) {
  return LABELS[key] ?? key.replace(/^[bp]_/, "").split("_").join(" ").toUpperCase();
}
