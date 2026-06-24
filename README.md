# GO Sports Analytics Scraper

A sports analytics scraper written in Go, with a React web UI. It collects MLB
player statistics from
[baseball-reference.com](https://www.baseball-reference.com/players/) and either
writes them to JSON/CSV files (CLI) or serves them to a browser (web app) where
you can search a player and view their last-N-days stats.

It covers two kinds of data:

- **Historical** — a player's bio plus season-by-season and career batting and
  pitching tables.
- **Recent / "live"** — a player's game-by-game log for a season (the freshest
  data baseball-reference exposes; the site is a reference database, not a true
  real-time feed).

Scraping is built on [colly](https://github.com/gocolly/colly).

## Install

Requires Go 1.24+.

```bash
git clone <this-repo>
cd GO-Sports-Analytics-Scraper
go build -o scraper ./cmd/scraper
```

This produces a `scraper` binary in the project root.

## Usage

### Find a player's ID

baseball-reference identifies players by an ID like `judgeaa01`. Look one up by
name:

```bash
./scraper search --name "Aaron Judge"
```

```
ID              NAME                        YEARS       ACTIVE
judgeaa01       Aaron Judge                 2016-2026   yes
```

### Historical stats (bio + season/career tables)

```bash
./scraper player --id judgeaa01 --format both
# or resolve by name automatically:
./scraper player --name "Jacob deGrom" --format json
```

Flags:

- `--id` — player ID (e.g. `judgeaa01`).
- `--name` — resolve a name to an ID instead of passing `--id`.
- `--format` — `json`, `csv`, or `both` (default `json`).
- `--out` — output directory (default `output`).

Outputs (under `output/`):

- `judgeaa01.json` — full record (bio + batting + pitching).
- `judgeaa01_batting.csv` — batting seasons + career row.
- `judgeaa01_pitching.csv` — pitching seasons + career row (pitchers only).

### Recent / "live" stats (season game log)

```bash
./scraper gamelog --id judgeaa01 --year 2026 --type batting --format both
./scraper gamelog --id degroja01 --year 2026 --type pitching
```

Flags:

- `--id` / `--name` — player.
- `--year` — season (default: current year).
- `--type` — `batting` or `pitching` (default `batting`).
- `--format` — `json`, `csv`, or `both`.
- `--out` — output directory.

## Web UI

A React app lets you search a player by name, pick a date window (default the
last 30 days), and see their aggregated stat line plus a game-by-game table. It
talks to a small Go HTTP API that wraps the scraper.

### 1. Start the API server

```bash
go build -o server ./cmd/server
./server            # listens on :8080
```

Endpoints:

- `GET /api/search?name=Aaron Judge` — resolve a name to player IDs.
- `GET /api/player?id=judgeaa01` — bio + season/career stats.
- `GET /api/recent?id=judgeaa01&type=batting&days=30&year=2026` — game logs
  filtered to the last N days plus an aggregated summary (counting stats summed,
  rate stats like AVG/OBP/SLG/OPS and ERA/WHIP recomputed).

### 2. Start the web app

```bash
cd web
npm install
npm run dev         # serves on http://localhost:5173
```

The dev server proxies `/api/*` to the Go server on `:8080`, so just open
http://localhost:5173 and search for a player.

> Note: if your environment can't reach the default npm registry, install with a
> mirror, e.g. `npm install --registry https://registry.npmmirror.com`.

The "last N days" window is measured from today's date, so an injured or
off-season player may show few or no games — widen the window or change the
season with the controls.

## Project layout

```
cmd/scraper/main.go        CLI (cobra): player, gamelog, search
cmd/server/main.go         HTTP+JSON API for the web UI (CORS enabled)
internal/scraper/
  collector.go             colly collector: rate limiting, retry, comment stripping
  parse.go                 shared helpers (IDs, URLs, row parsing)
  player.go                bio + season/career batting & pitching tables
  gamelog.go               season game-by-game logs
  index.go                 name -> player ID search
internal/models/models.go  Player, SeasonBatting, SeasonPitching, GameLog
internal/stats/aggregate.go  date-window filtering + batting/pitching aggregation
internal/output/writer.go  JSON and CSV writers
web/                       React + Vite + TypeScript front end
```

## How it handles baseball-reference

- **Commented tables.** baseball-reference hides most stat tables inside HTML
  comments (`<!-- ... -->`) to deter scrapers. The collector strips comment
  markers in an `OnResponse` hook so the tables become real DOM nodes that
  colly's `OnHTML` selectors can match.
- **Stable column keys.** Cells are read by their `data-stat` attribute (e.g.
  `b_hr`, `p_era`) rather than column position, which survives layout changes.

## Please scrape responsibly

baseball-reference rate-limits aggressively (roughly 20 requests/minute) and
will return HTTP 429 if you exceed it. This tool is deliberately polite:

- one request at a time, with a delay plus jitter between requests;
- a descriptive `User-Agent`;
- exponential-backoff retries on errors (including 429).

Use this for personal/educational analytics, respect baseball-reference's
[terms of use](https://www.sports-reference.com/termsofuse.html), and don't
remove the rate limiting.
