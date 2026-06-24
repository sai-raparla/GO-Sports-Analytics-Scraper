// Package stats filters and aggregates scraped game logs into summary stat
// lines (e.g. "last 30 days").
package stats

import (
	"fmt"
	"strconv"
	"time"

	"sports-scraper/internal/models"
)

// DateLayout is the date format baseball-reference uses in game logs.
const DateLayout = "2006-01-02"

// FilterByDays returns the subset of logs whose date falls within the last
// `days` days, counting back from `now` (inclusive). Logs with unparseable
// dates are dropped.
func FilterByDays(logs []models.GameLog, days int, now time.Time) []models.GameLog {
	cutoff := now.AddDate(0, 0, -days)
	out := make([]models.GameLog, 0, len(logs))
	for _, l := range logs {
		d, err := time.Parse(DateLayout, l.Date)
		if err != nil {
			continue
		}
		if !d.Before(cutoff) && !d.After(now) {
			out = append(out, l)
		}
	}
	return out
}

// BattingSummary is an aggregated batting stat line.
type BattingSummary struct {
	Games   int     `json:"games"`
	PA      int     `json:"pa"`
	AB      int     `json:"ab"`
	R       int     `json:"r"`
	H       int     `json:"h"`
	Doubles int     `json:"2b"`
	Triples int     `json:"3b"`
	HR      int     `json:"hr"`
	RBI     int     `json:"rbi"`
	BB      int     `json:"bb"`
	SO      int     `json:"so"`
	SB      int     `json:"sb"`
	CS      int     `json:"cs"`
	HBP     int     `json:"hbp"`
	SF      int     `json:"sf"`
	TB      int     `json:"tb"`
	AVG     float64 `json:"avg"`
	OBP     float64 `json:"obp"`
	SLG     float64 `json:"slg"`
	OPS     float64 `json:"ops"`
}

// AggregateBatting sums the batting game logs and recomputes rate stats.
func AggregateBatting(logs []models.GameLog) BattingSummary {
	var s BattingSummary
	s.Games = len(logs)
	for _, l := range logs {
		s.PA += atoi(l.Stats["b_pa"])
		s.AB += atoi(l.Stats["b_ab"])
		s.R += atoi(l.Stats["b_r"])
		s.H += atoi(l.Stats["b_h"])
		s.Doubles += atoi(l.Stats["b_doubles"])
		s.Triples += atoi(l.Stats["b_triples"])
		s.HR += atoi(l.Stats["b_hr"])
		s.RBI += atoi(l.Stats["b_rbi"])
		s.BB += atoi(l.Stats["b_bb"])
		s.SO += atoi(l.Stats["b_so"])
		s.SB += atoi(l.Stats["b_sb"])
		s.CS += atoi(l.Stats["b_cs"])
		s.HBP += atoi(l.Stats["b_hbp"])
		s.SF += atoi(l.Stats["b_sf"])
		s.TB += atoi(l.Stats["b_tb"])
	}
	// Fall back to computing total bases if the column was missing.
	if s.TB == 0 {
		singles := s.H - s.Doubles - s.Triples - s.HR
		s.TB = singles + 2*s.Doubles + 3*s.Triples + 4*s.HR
	}

	if s.AB > 0 {
		s.AVG = round3(float64(s.H) / float64(s.AB))
		s.SLG = round3(float64(s.TB) / float64(s.AB))
	}
	if obpDen := s.AB + s.BB + s.HBP + s.SF; obpDen > 0 {
		s.OBP = round3(float64(s.H+s.BB+s.HBP) / float64(obpDen))
	}
	s.OPS = round3(s.OBP + s.SLG)
	return s
}

// PitchingSummary is an aggregated pitching stat line.
type PitchingSummary struct {
	Games int     `json:"games"`
	Outs  int     `json:"outs"`
	IP    string  `json:"ip"`
	H     int     `json:"h"`
	R     int     `json:"r"`
	ER    int     `json:"er"`
	BB    int     `json:"bb"`
	SO    int     `json:"so"`
	HR    int     `json:"hr"`
	ERA   float64 `json:"era"`
	WHIP  float64 `json:"whip"`
	K9    float64 `json:"k9"`
	BB9   float64 `json:"bb9"`
}

// AggregatePitching sums the pitching game logs and recomputes rate stats.
// Innings pitched are summed in outs to handle the .1/.2 (thirds) notation.
func AggregatePitching(logs []models.GameLog) PitchingSummary {
	var s PitchingSummary
	s.Games = len(logs)
	for _, l := range logs {
		s.Outs += ipToOuts(l.Stats["p_ip"])
		s.H += atoi(l.Stats["p_h"])
		s.R += atoi(l.Stats["p_r"])
		s.ER += atoi(l.Stats["p_er"])
		s.BB += atoi(l.Stats["p_bb"])
		s.SO += atoi(l.Stats["p_so"])
		s.HR += atoi(l.Stats["p_hr"])
	}
	s.IP = outsToIP(s.Outs)
	if s.Outs > 0 {
		ip := float64(s.Outs) / 3.0
		s.ERA = round2(9 * float64(s.ER) / ip)
		s.WHIP = round2(float64(s.BB+s.H) / ip)
		s.K9 = round2(9 * float64(s.SO) / ip)
		s.BB9 = round2(9 * float64(s.BB) / ip)
	}
	return s
}

// ipToOuts converts a baseball "innings pitched" value like "6.1" (6 and 1/3)
// into a number of outs.
func ipToOuts(ip string) int {
	if ip == "" {
		return 0
	}
	whole := ip
	frac := 0
	if dot := indexByte(ip, '.'); dot >= 0 {
		whole = ip[:dot]
		switch ip[dot+1:] {
		case "1":
			frac = 1
		case "2":
			frac = 2
		}
	}
	return atoi(whole)*3 + frac
}

func outsToIP(outs int) string {
	return fmt.Sprintf("%d.%d", outs/3, outs%3)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func round3(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }
func round2(f float64) float64 { return float64(int(f*100+0.5)) / 100 }
