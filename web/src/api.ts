import type { Player, RecentResponse, SearchResult, StatType } from "./types";

async function getJSON<T>(url: string, signal?: AbortSignal): Promise<T> {
  const res = await fetch(url, { signal });
  if (!res.ok) {
    let msg = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      if (body?.error) msg = body.error;
    } catch {
      // ignore non-JSON error bodies
    }
    throw new Error(msg);
  }
  return res.json() as Promise<T>;
}

export function searchPlayers(name: string, signal?: AbortSignal) {
  return getJSON<SearchResult[]>(
    `/api/search?name=${encodeURIComponent(name)}`,
    signal,
  );
}

export function fetchPlayer(id: string, signal?: AbortSignal) {
  return getJSON<Player>(`/api/player?id=${encodeURIComponent(id)}`, signal);
}

export function fetchRecent(
  id: string,
  type: StatType,
  days: number,
  year: number,
  signal?: AbortSignal,
) {
  const q = new URLSearchParams({
    id,
    type,
    days: String(days),
    year: String(year),
  });
  return getJSON<RecentResponse>(`/api/recent?${q.toString()}`, signal);
}
