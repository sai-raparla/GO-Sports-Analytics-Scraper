// Command server exposes the scraper over a small HTTP+JSON API for the web UI.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"sports-scraper/internal/models"
	"sports-scraper/internal/scraper"
	"sports-scraper/internal/stats"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", handleSearch)
	mux.HandleFunc("/api/player", handlePlayer)
	mux.HandleFunc("/api/team", handleTeam)
	mux.HandleFunc("/api/recent", handleRecent)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	log.Printf("scraper API listening on %s", *addr)
	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// handleSearch resolves a (partial) player name to candidate players.
func handleSearch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing 'name' query parameter")
		return
	}
	results, err := scraper.SearchPlayers(name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// handlePlayer returns a player's bio + season/career stats.
func handlePlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing 'id' query parameter")
		return
	}
	player, err := scraper.FetchPlayer(id)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, player)
}

// handleTeam returns a team's season batting and pitching totals.
func handleTeam(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	teamQuery := q.Get("team")
	if teamQuery == "" {
		teamQuery = q.Get("id")
	}
	if teamQuery == "" {
		teamQuery = q.Get("name")
	}
	if teamQuery == "" {
		writeError(w, http.StatusBadRequest, "missing 'team' query parameter")
		return
	}

	year := atoiDefault(q.Get("year"), time.Now().Year())
	team, err := scraper.FetchTeam(teamQuery, year)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, team)
}

type recentResponse struct {
	PlayerID string                 `json:"player_id"`
	Type     string                 `json:"type"`
	Year     int                    `json:"year"`
	Days     int                    `json:"days"`
	From     string                 `json:"from"`
	To       string                 `json:"to"`
	Games    int                    `json:"games"`
	Batting  *stats.BattingSummary  `json:"batting,omitempty"`
	Pitching *stats.PitchingSummary `json:"pitching,omitempty"`
	GameLogs []models.GameLog       `json:"game_logs"`
}

// handleRecent returns a player's game logs filtered to the last N days, plus
// an aggregated summary stat line.
func handleRecent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id := q.Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing 'id' query parameter")
		return
	}
	logType := q.Get("type")
	if logType == "" {
		logType = "batting"
	}
	year := atoiDefault(q.Get("year"), time.Now().Year())
	days := atoiDefault(q.Get("days"), 30)

	logs, err := scraper.FetchGameLog(id, logType, year)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	now := time.Now()

	var filtered []models.GameLog
	var from, to string
	respDays := days
	// Apply the rolling "last N days" window only for the current season and
	// only when a positive window was requested. days == 0 means "full season",
	// and past seasons are always returned in full (the window is meaningless
	// relative to today).
	if days > 0 && year == now.Year() {
		filtered = stats.FilterByDays(logs, days, now)
		from = now.AddDate(0, 0, -days).Format(stats.DateLayout)
		to = now.Format(stats.DateLayout)
	} else {
		filtered = logs
		respDays = 0
		from, to = logDateRange(logs)
	}

	resp := recentResponse{
		PlayerID: id,
		Type:     logType,
		Year:     year,
		Days:     respDays,
		From:     from,
		To:       to,
		Games:    len(filtered),
		GameLogs: filtered,
	}
	switch logType {
	case "pitching", "p":
		agg := stats.AggregatePitching(filtered)
		resp.Pitching = &agg
		resp.Type = "pitching"
	default:
		agg := stats.AggregateBatting(filtered)
		resp.Batting = &agg
		resp.Type = "batting"
	}

	writeJSON(w, http.StatusOK, resp)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// logDateRange returns the earliest and latest game dates in logs. Dates use
// the sortable YYYY-MM-DD layout, so lexical comparison matches chronological
// order. Empty strings are returned when logs is empty.
func logDateRange(logs []models.GameLog) (from, to string) {
	for _, l := range logs {
		if l.Date == "" {
			continue
		}
		if from == "" || l.Date < from {
			from = l.Date
		}
		if to == "" || l.Date > to {
			to = l.Date
		}
	}
	return from, to
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
