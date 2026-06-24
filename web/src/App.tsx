import { useEffect, useState } from "react";
import { fetchPlayer, fetchRecent, fetchTeam } from "./api";
import { SearchBox } from "./components/SearchBox";
import { PlayerHeader } from "./components/PlayerHeader";
import { SummaryCards } from "./components/SummaryCards";
import { GameLogTable } from "./components/GameLogTable";
import { TeamSearch } from "./components/TeamSearch";
import { TeamStatsPanel } from "./components/TeamStatsPanel";
import type { Player, RecentResponse, SearchResult, StatType, Team } from "./types";

const DAY_OPTIONS = [7, 14, 30, 60, 90];
const CURRENT_YEAR = new Date().getFullYear();
const YEAR_OPTIONS = Array.from({ length: 6 }, (_, i) => CURRENT_YEAR - i);

export function App() {
  const [mode, setMode] = useState<"player" | "team">("player");
  const [selected, setSelected] = useState<SearchResult | null>(null);
  const [player, setPlayer] = useState<Player | null>(null);
  const [recent, setRecent] = useState<RecentResponse | null>(null);
  const [type, setType] = useState<StatType>("batting");
  const [days, setDays] = useState(30);
  const [year, setYear] = useState(CURRENT_YEAR);
  const [teamQuery, setTeamQuery] = useState("");
  const [teamYear, setTeamYear] = useState(CURRENT_YEAR);
  const [teamType, setTeamType] = useState<StatType>("batting");
  const [team, setTeam] = useState<Team | null>(null);
  const [teamLoading, setTeamLoading] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isCurrentSeason = year === CURRENT_YEAR;

  // Fetch bio once per selected player.
  useEffect(() => {
    if (!selected) return;
    const controller = new AbortController();
    fetchPlayer(selected.id, controller.signal)
      .then(setPlayer)
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      });
    return () => controller.abort();
  }, [selected]);

  // Fetch the filtered recent stats whenever the player or filters change.
  useEffect(() => {
    if (!selected) return;
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    fetchRecent(selected.id, type, days, year, controller.signal)
      .then(setRecent)
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, [selected, type, days, year]);

  useEffect(() => {
    if (!teamQuery) return;
    const controller = new AbortController();
    setTeamLoading(true);
    setError(null);
    fetchTeam(teamQuery, teamYear, controller.signal)
      .then(setTeam)
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      })
      .finally(() => setTeamLoading(false));
    return () => controller.abort();
  }, [teamQuery, teamYear]);

  function onSelect(p: SearchResult) {
    setSelected(p);
    setPlayer(null);
    setRecent(null);
    setError(null);
  }

  function onTeamSubmit(query: string) {
    setTeamQuery(query);
    setTeam(null);
    setError(null);
  }

  return (
    <div className="app">
      <div className="hero">
        <h1>
          <span className="hero-accent">Baseball</span> Stats Explorer
        </h1>
        <p className="hero-sub">
          Search player game logs or team season totals.
        </p>
        <div className="mode-switch" role="tablist" aria-label="Search mode">
          {(["player", "team"] as const).map((m) => (
            <button
              key={m}
              className={`toggle-btn ${mode === m ? "active" : ""}`}
              onClick={() => {
                setMode(m);
                setError(null);
              }}
            >
              {m[0].toUpperCase() + m.slice(1)}
            </button>
          ))}
        </div>
        {mode === "player" ? (
          <SearchBox onSelect={onSelect} />
        ) : (
          <TeamSearch
            loading={teamLoading}
            year={teamYear}
            yearOptions={YEAR_OPTIONS}
            onYearChange={setTeamYear}
            onSubmit={onTeamSubmit}
          />
        )}
      </div>

      {mode === "player" && selected && (
        <main className="content">
          {player && <PlayerHeader player={player} />}

          <div className="controls">
            <div className="toggle">
              {(["batting", "pitching"] as StatType[]).map((t) => (
                <button
                  key={t}
                  className={`toggle-btn ${type === t ? "active" : ""}`}
                  onClick={() => setType(t)}
                >
                  {t[0].toUpperCase() + t.slice(1)}
                </button>
              ))}
            </div>

            <label className="control">
              <span>Window</span>
              <select
                value={isCurrentSeason ? days : 0}
                disabled={!isCurrentSeason}
                onChange={(e) => setDays(Number(e.target.value))}
              >
                <option value={0}>Full season</option>
                {DAY_OPTIONS.map((d) => (
                  <option key={d} value={d}>
                    Last {d} days
                  </option>
                ))}
              </select>
            </label>

            <label className="control">
              <span>Season</span>
              <select
                value={year}
                onChange={(e) => setYear(Number(e.target.value))}
              >
                {YEAR_OPTIONS.map((y) => (
                  <option key={y} value={y}>
                    {y}
                  </option>
                ))}
              </select>
            </label>
          </div>

          {error && <div className="error-banner">{error}</div>}
          {loading && <div className="loading">Loading stats…</div>}

          {recent && !loading && (
            <>
              <SummaryCards recent={recent} />
              <section className="games-section">
                <h3>Game log</h3>
                <GameLogTable logs={recent.game_logs} type={type} />
              </section>
            </>
          )}
        </main>
      )}

      {mode === "team" && (
        <>
          {error && <div className="error-banner">{error}</div>}
          {teamLoading && <div className="loading">Loading team stats...</div>}
          {team && !teamLoading && (
            <TeamStatsPanel team={team} type={teamType} onTypeChange={setTeamType} />
          )}
        </>
      )}

      {mode === "player" && !selected && (
        <div className="placeholder">
          <p>Start by searching for a player above.</p>
        </div>
      )}

      {mode === "team" && !teamQuery && (
        <div className="placeholder">
          <p>Start by entering a team above.</p>
        </div>
      )}

      <footer className="footer">
        Data scraped from baseball-reference.com · for educational use
      </footer>
    </div>
  );
}
