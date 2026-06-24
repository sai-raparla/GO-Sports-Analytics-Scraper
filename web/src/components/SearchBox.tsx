import { useEffect, useRef, useState } from "react";
import { searchPlayers } from "../api";
import { useDebounced } from "../hooks";
import type { SearchResult } from "../types";

interface Props {
  onSelect: (player: SearchResult) => void;
}

export function SearchBox({ onSelect }: Props) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debounced = useDebounced(query.trim(), 350);
  const boxRef = useRef<HTMLDivElement>(null);
  // Name of the player just picked. Selecting fills the input with the full
  // name, which would otherwise re-trigger the search and reopen the dropdown.
  const selectedName = useRef<string | null>(null);

  useEffect(() => {
    if (debounced.length < 2) {
      setResults([]);
      setError(null);
      return;
    }
    if (debounced === selectedName.current) {
      return;
    }
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    searchPlayers(debounced, controller.signal)
      .then((r) => {
        setResults(r);
        setOpen(true);
      })
      .catch((e) => {
        if (e.name !== "AbortError") setError(e.message);
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, [debounced]);

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onClickOutside);
    return () => document.removeEventListener("mousedown", onClickOutside);
  }, []);

  function pick(p: SearchResult) {
    selectedName.current = p.name;
    setQuery(p.name);
    setResults([]);
    setOpen(false);
    onSelect(p);
  }

  return (
    <div className="searchbox" ref={boxRef}>
      <input
        className="search-input"
        type="text"
        value={query}
        placeholder="Search a player, e.g. Aaron Judge"
        onChange={(e) => {
          selectedName.current = null;
          setQuery(e.target.value);
        }}
        onFocus={() => results.length > 0 && setOpen(true)}
        autoFocus
      />
      {loading && <span className="search-spinner" />}
      {open && (results.length > 0 || error) && (
        <ul className="search-results">
          {error && <li className="search-error">{error}</li>}
          {results.map((r) => (
            <li key={r.id}>
              <button className="search-result" onClick={() => pick(r)}>
                <span className="result-name">{r.name}</span>
                <span className="result-meta">
                  {r.years || ""}
                  {r.active && <span className="badge-active">active</span>}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
