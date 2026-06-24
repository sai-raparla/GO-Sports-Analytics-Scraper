# GO Sports Analytics Scraper

A command-line sports analytics scraper written in Go. It collects MLB player
and team statistics from [baseball-reference.com](https://www.baseball-reference.com/)
and writes them to JSON/CSV files.

It covers three kinds of data:

- **Historical** — a player's bio plus season-by-season and career batting and
  pitching tables.
- **Team seasons** — a team's batting and pitching totals for a season.
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

### Team season stats

```bash
./scraper team --id NYY --year 2026 --format both
./scraper team --id LAD --year 2026 --format json
```

Flags:

- `--id` — baseball-reference team ID (e.g. `NYY`, `LAD`, `BOS`).
- `--year` — season (default: current year).
- `--format` — `json`, `csv`, or `both` (default `json`).
- `--out` — output directory (default `output`).

Outputs (under `output/`):

- `NYY_2026.json` — full team totals record.
- `NYY_2026_batting.csv` — one row of team batting totals.
- `NYY_2026_pitching.csv` — one row of team pitching totals.

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

## Project layout

```
cmd/scraper/main.go        CLI (cobra): player, team, gamelog, search
internal/scraper/
  collector.go             colly collector: rate limiting, retry, comment stripping
  parse.go                 shared helpers (IDs, URLs, row parsing)
  player.go                bio + season/career batting & pitching tables
  team.go                  team batting & pitching totals
  gamelog.go               season game-by-game logs
  index.go                 name -> player ID search
internal/models/models.go  Player, Team, SeasonBatting, SeasonPitching, GameLog
internal/output/writer.go  JSON and CSV writers
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
