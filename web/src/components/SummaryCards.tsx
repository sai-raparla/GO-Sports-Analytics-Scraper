import type { RecentResponse } from "../types";

interface Props {
  recent: RecentResponse;
}

const fmt3 = (n: number) => n.toFixed(3).replace(/^0/, "");
const fmt2 = (n: number) => n.toFixed(2);

export function SummaryCards({ recent }: Props) {
  let cards: { label: string; value: string }[] = [];

  if (recent.type === "batting" && recent.batting) {
    const b = recent.batting;
    cards = [
      { label: "AVG", value: fmt3(b.avg) },
      { label: "OBP", value: fmt3(b.obp) },
      { label: "SLG", value: fmt3(b.slg) },
      { label: "OPS", value: fmt3(b.ops) },
      { label: "HR", value: String(b.hr) },
      { label: "RBI", value: String(b.rbi) },
      { label: "H", value: String(b.h) },
      { label: "R", value: String(b.r) },
      { label: "BB", value: String(b.bb) },
      { label: "SO", value: String(b.so) },
      { label: "SB", value: String(b.sb) },
      { label: "PA", value: String(b.pa) },
    ];
  } else if (recent.type === "pitching" && recent.pitching) {
    const p = recent.pitching;
    cards = [
      { label: "ERA", value: fmt2(p.era) },
      { label: "WHIP", value: fmt2(p.whip) },
      { label: "IP", value: p.ip },
      { label: "SO", value: String(p.so) },
      { label: "BB", value: String(p.bb) },
      { label: "H", value: String(p.h) },
      { label: "ER", value: String(p.er) },
      { label: "HR", value: String(p.hr) },
      { label: "K/9", value: fmt2(p.k9) },
      { label: "BB/9", value: fmt2(p.bb9) },
    ];
  }

  return (
    <section className="summary">
      <div className="summary-heading">
        <h3>
          Last {recent.days} days
          <span className="summary-window">
            {recent.from} &rarr; {recent.to}
          </span>
        </h3>
        <span className="summary-games">{recent.games} games</span>
      </div>
      <div className="stat-grid">
        {cards.map((c) => (
          <div className="stat-card" key={c.label}>
            <div className="stat-value">{c.value}</div>
            <div className="stat-label">{c.label}</div>
          </div>
        ))}
      </div>
    </section>
  );
}
