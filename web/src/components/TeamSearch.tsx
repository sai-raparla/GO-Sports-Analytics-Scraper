import { FormEvent, useEffect, useMemo, useRef, useState } from "react";

interface TeamOption {
  id: string;
  name: string;
}

const TEAMS: TeamOption[] = [
  { id: "ARI", name: "Arizona Diamondbacks" },
  { id: "ATL", name: "Atlanta Braves" },
  { id: "BAL", name: "Baltimore Orioles" },
  { id: "BOS", name: "Boston Red Sox" },
  { id: "CHC", name: "Chicago Cubs" },
  { id: "CHW", name: "Chicago White Sox" },
  { id: "CIN", name: "Cincinnati Reds" },
  { id: "CLE", name: "Cleveland Guardians" },
  { id: "COL", name: "Colorado Rockies" },
  { id: "DET", name: "Detroit Tigers" },
  { id: "HOU", name: "Houston Astros" },
  { id: "KCR", name: "Kansas City Royals" },
  { id: "LAA", name: "Los Angeles Angels" },
  { id: "LAD", name: "Los Angeles Dodgers" },
  { id: "MIA", name: "Miami Marlins" },
  { id: "MIL", name: "Milwaukee Brewers" },
  { id: "MIN", name: "Minnesota Twins" },
  { id: "NYM", name: "New York Mets" },
  { id: "NYY", name: "New York Yankees" },
  { id: "ATH", name: "Athletics" },
  { id: "PHI", name: "Philadelphia Phillies" },
  { id: "PIT", name: "Pittsburgh Pirates" },
  { id: "SDP", name: "San Diego Padres" },
  { id: "SFG", name: "San Francisco Giants" },
  { id: "SEA", name: "Seattle Mariners" },
  { id: "STL", name: "St. Louis Cardinals" },
  { id: "TBR", name: "Tampa Bay Rays" },
  { id: "TEX", name: "Texas Rangers" },
  { id: "TOR", name: "Toronto Blue Jays" },
  { id: "WSN", name: "Washington Nationals" },
];

interface Props {
  loading: boolean;
  year: number;
  yearOptions: number[];
  onYearChange: (year: number) => void;
  onSubmit: (team: string) => void;
}

export function TeamSearch({
  loading,
  year,
  yearOptions,
  onYearChange,
  onSubmit,
}: Props) {
  const [team, setTeam] = useState("");
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);

  const suggestions = useMemo(() => {
    const q = team.trim().toLowerCase();
    if (q.length < 1) return [];
    return TEAMS.filter((t) => {
      const name = t.name.toLowerCase();
      const parts = name.split(" ");
      const nickname = parts[parts.length - 1] ?? "";
      return name.includes(q) || nickname.includes(q) || t.id.toLowerCase().includes(q);
    }).slice(0, 8);
  }, [team]);

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onClickOutside);
    return () => document.removeEventListener("mousedown", onClickOutside);
  }, []);

  function submit(e: FormEvent) {
    e.preventDefault();
    const trimmed = team.trim();
    if (trimmed) onSubmit(trimmed);
  }

  function pick(option: TeamOption) {
    setTeam(option.name);
    setOpen(false);
    onSubmit(option.id);
  }

  return (
    <form className="team-search" onSubmit={submit}>
      <div className="team-searchbox" ref={boxRef}>
        <input
          className="search-input"
          type="text"
          value={team}
          placeholder="Team name or ID, e.g. Yankees"
          onChange={(e) => {
            setTeam(e.target.value);
            setOpen(true);
          }}
          onFocus={() => suggestions.length > 0 && setOpen(true)}
        />
        {open && suggestions.length > 0 && (
          <ul className="search-results">
            {suggestions.map((option) => (
              <li key={option.id}>
                <button
                  className="search-result"
                  type="button"
                  onClick={() => pick(option)}
                >
                  <span className="result-name">{option.name}</span>
                  <span className="result-meta">{option.id}</span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
      <label className="control">
        <span>Season</span>
        <select value={year} onChange={(e) => onYearChange(Number(e.target.value))}>
          {yearOptions.map((y) => (
            <option key={y} value={y}>
              {y}
            </option>
          ))}
        </select>
      </label>
      <button className="primary-btn" type="submit" disabled={loading || !team.trim()}>
        {loading ? "Loading" : "Load"}
      </button>
    </form>
  );
}
