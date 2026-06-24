// Package output writes scraped data to disk as JSON or CSV.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"sports-scraper/internal/models"
)

// EnsureDir creates dir (and parents) if needed.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// WriteJSON marshals v to path as indented JSON.
func WriteJSON(path string, v any) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WritePlayerCSV writes a player's batting and pitching season tables to two
// CSV files (only those that have rows). It returns the paths written.
func WritePlayerCSV(dir, id string, p *models.Player) ([]string, error) {
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}
	var written []string

	if len(p.Batting) > 0 {
		path := filepath.Join(dir, id+"_batting.csv")
		header := []string{"season", "age", "team", "league", "war", "g", "pa", "ab", "r", "h", "2b", "3b", "hr", "rbi", "sb", "cs", "bb", "so", "ba", "obp", "slg", "ops"}
		rows := make([][]string, 0, len(p.Batting))
		for _, b := range p.Batting {
			rows = append(rows, []string{b.Season, b.Age, b.Team, b.League, b.WAR, b.G, b.PA, b.AB, b.R, b.H, b.Doubles, b.Triples, b.HR, b.RBI, b.SB, b.CS, b.BB, b.SO, b.BA, b.OBP, b.SLG, b.OPS})
		}
		if err := writeCSV(path, header, rows); err != nil {
			return written, err
		}
		written = append(written, path)
	}

	if len(p.Pitching) > 0 {
		path := filepath.Join(dir, id+"_pitching.csv")
		header := []string{"season", "age", "team", "league", "war", "w", "l", "era", "g", "gs", "sv", "ip", "h", "r", "er", "bb", "so", "whip"}
		rows := make([][]string, 0, len(p.Pitching))
		for _, s := range p.Pitching {
			rows = append(rows, []string{s.Season, s.Age, s.Team, s.League, s.WAR, s.W, s.L, s.ERA, s.G, s.GS, s.SV, s.IP, s.H, s.R, s.ER, s.BB, s.SO, s.WHIP})
		}
		if err := writeCSV(path, header, rows); err != nil {
			return written, err
		}
		written = append(written, path)
	}

	return written, nil
}

// WriteGameLogCSV writes game logs to a CSV. Stat columns are the union of all
// stat keys, sorted for stable output.
func WriteGameLogCSV(path string, logs []models.GameLog) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	statKeys := map[string]struct{}{}
	for _, l := range logs {
		for k := range l.Stats {
			statKeys[k] = struct{}{}
		}
	}
	sortedStats := make([]string, 0, len(statKeys))
	for k := range statKeys {
		sortedStats = append(sortedStats, k)
	}
	sort.Strings(sortedStats)

	header := append([]string{"date", "team", "opponent"}, sortedStats...)
	rows := make([][]string, 0, len(logs))
	for _, l := range logs {
		row := []string{l.Date, l.Team, l.Opponent}
		for _, k := range sortedStats {
			row = append(row, l.Stats[k])
		}
		rows = append(rows, row)
	}
	return writeCSV(path, header, rows)
}

// WriteTeamCSV writes a team's batting and pitching totals to separate CSV
// files. Stat columns are sorted for stable output.
func WriteTeamCSV(dir string, t *models.Team) ([]string, error) {
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}

	var written []string
	if len(t.BattingTotals) > 0 {
		path := filepath.Join(dir, fmt.Sprintf("%s_%d_batting.csv", t.ID, t.Year))
		if err := writeTeamTotalsCSV(path, t, t.BattingTotals); err != nil {
			return written, err
		}
		written = append(written, path)
	}
	if len(t.PitchingTotals) > 0 {
		path := filepath.Join(dir, fmt.Sprintf("%s_%d_pitching.csv", t.ID, t.Year))
		if err := writeTeamTotalsCSV(path, t, t.PitchingTotals); err != nil {
			return written, err
		}
		written = append(written, path)
	}
	return written, nil
}

func writeTeamTotalsCSV(path string, t *models.Team, stats map[string]string) error {
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	header := append([]string{"team_id", "year", "team_name"}, keys...)
	row := []string{t.ID, fmt.Sprint(t.Year), t.Name}
	for _, k := range keys {
		row = append(row, stats[k])
	}
	return writeCSV(path, header, [][]string{row})
}

func writeCSV(path string, header []string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("writing csv %s: %w", path, err)
	}
	return nil
}
