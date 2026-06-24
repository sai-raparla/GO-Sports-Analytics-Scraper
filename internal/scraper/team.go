package scraper

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	"sports-scraper/internal/models"
)

var teamIDRe = regexp.MustCompile(`^[A-Za-z]{2,3}$`)

var teamNameToID = map[string]string{
	"arizona diamondbacks":   "ARI",
	"diamondbacks":           "ARI",
	"atlanta braves":         "ATL",
	"braves":                 "ATL",
	"baltimore orioles":      "BAL",
	"orioles":                "BAL",
	"boston red sox":         "BOS",
	"red sox":                "BOS",
	"chicago cubs":           "CHC",
	"cubs":                   "CHC",
	"chicago white sox":      "CHW",
	"white sox":              "CHW",
	"cincinnati reds":        "CIN",
	"reds":                   "CIN",
	"cleveland guardians":    "CLE",
	"guardians":              "CLE",
	"cleveland indians":      "CLE",
	"indians":                "CLE",
	"colorado rockies":       "COL",
	"rockies":                "COL",
	"detroit tigers":         "DET",
	"tigers":                 "DET",
	"houston astros":         "HOU",
	"astros":                 "HOU",
	"kansas city royals":     "KCR",
	"royals":                 "KCR",
	"los angeles angels":     "LAA",
	"la angels":              "LAA",
	"angels":                 "LAA",
	"los angeles dodgers":    "LAD",
	"la dodgers":             "LAD",
	"dodgers":                "LAD",
	"miami marlins":          "MIA",
	"florida marlins":        "FLA",
	"marlins":                "MIA",
	"milwaukee brewers":      "MIL",
	"brewers":                "MIL",
	"minnesota twins":        "MIN",
	"twins":                  "MIN",
	"new york mets":          "NYM",
	"mets":                   "NYM",
	"new york yankees":       "NYY",
	"yankees":                "NYY",
	"oakland athletics":      "OAK",
	"athletics":              "ATH",
	"philadelphia athletics": "PHA",
	"philadelphia phillies":  "PHI",
	"phillies":               "PHI",
	"pittsburgh pirates":     "PIT",
	"pirates":                "PIT",
	"san diego padres":       "SDP",
	"padres":                 "SDP",
	"san francisco giants":   "SFG",
	"giants":                 "SFG",
	"seattle mariners":       "SEA",
	"mariners":               "SEA",
	"st louis cardinals":     "STL",
	"st. louis cardinals":    "STL",
	"cardinals":              "STL",
	"tampa bay rays":         "TBR",
	"tampa bay devil rays":   "TBD",
	"rays":                   "TBR",
	"texas rangers":          "TEX",
	"rangers":                "TEX",
	"toronto blue jays":      "TOR",
	"blue jays":              "TOR",
	"washington nationals":   "WSN",
	"montreal expos":         "MON",
	"nationals":              "WSN",
}

// ResolveTeamID resolves an MLB team abbreviation or common team name to the
// baseball-reference team ID used in team season URLs.
func ResolveTeamID(team string) (string, error) {
	team = strings.TrimSpace(team)
	if team == "" {
		return "", fmt.Errorf("empty team")
	}
	if teamIDRe.MatchString(team) {
		return strings.ToUpper(team), nil
	}

	key := normalizeTeamName(team)
	if id, ok := teamNameToID[key]; ok {
		return id, nil
	}

	var matches []string
	for name, id := range teamNameToID {
		if strings.Contains(name, key) || strings.Contains(key, name) {
			matches = append(matches, fmt.Sprintf("%s (%s)", name, id))
		}
	}
	sort.Strings(matches)
	if len(matches) == 1 {
		parts := strings.Fields(matches[0])
		return strings.Trim(parts[len(parts)-1], "()"), nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("%q matched multiple teams: %s", team, strings.Join(matches, ", "))
	}
	return "", fmt.Errorf("unknown team %q; use an abbreviation like NYY or a full team name", team)
}

// FetchTeam scrapes one team's batting and pitching totals for a season.
func FetchTeam(team string, year int) (*models.Team, error) {
	id, err := ResolveTeamID(team)
	if err != nil {
		return nil, err
	}

	c := newCollector()
	url := teamURL(id, year)
	t := &models.Team{
		ID:        id,
		Year:      year,
		Name:      teamDisplayName(id),
		SourceURL: url,
	}

	c.OnHTML("#meta h1", func(e *colly.HTMLElement) {
		if name := parseTeamPageName(e.DOM.Text(), year); name != "" {
			t.Name = name
		}
	})
	c.OnHTML("#players_standard_batting", func(e *colly.HTMLElement) {
		t.BattingTotals = totalsFromTable(e.DOM, "b_")
	})
	c.OnHTML("#players_standard_pitching", func(e *colly.HTMLElement) {
		t.PitchingTotals = totalsFromTable(e.DOM, "p_")
	})

	var visitErr error
	c.OnError(func(r *colly.Response, err error) {
		if getRetryCount(r) >= maxRetries {
			visitErr = fmt.Errorf("request to %s failed: %w", r.Request.URL, err)
		}
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}
	c.Wait()

	if visitErr != nil {
		return nil, visitErr
	}
	if len(t.BattingTotals) == 0 && len(t.PitchingTotals) == 0 {
		return nil, fmt.Errorf("no data found for team %s in %d (does the page exist?)", id, year)
	}
	return t, nil
}

func normalizeTeamName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ".", "")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func totalsFromTable(table *goquery.Selection, prefix string) map[string]string {
	var totals map[string]string
	table.Find("tfoot tr").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		cells := rowCells(row)
		filtered := filterStatPrefix(cells, prefix)
		if len(filtered) == 0 {
			return true
		}
		totals = filtered
		return false
	})
	if totals != nil {
		return totals
	}

	table.Find("tbody tr").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		if isHeaderRow(row) {
			return true
		}
		label := strings.ToLower(firstCellText(row))
		if !strings.Contains(label, "team total") {
			return true
		}
		totals = filterStatPrefix(rowCells(row), prefix)
		return false
	})
	return totals
}

func filterStatPrefix(cells map[string]string, prefix string) map[string]string {
	out := make(map[string]string)
	for k, v := range cells {
		if strings.HasPrefix(k, prefix) && v != "" {
			out[k] = v
		}
	}
	return out
}

func parseTeamPageName(title string, year int) string {
	title = strings.Join(strings.Fields(title), " ")
	title = strings.TrimPrefix(title, fmt.Sprintf("%d ", year))
	title = strings.TrimSuffix(title, " Statistics")
	return strings.TrimSpace(title)
}

func teamDisplayName(id string) string {
	for name, mappedID := range teamNameToID {
		if mappedID == id && strings.Contains(name, " ") {
			parts := strings.Fields(name)
			for i, p := range parts {
				parts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
			return strings.Join(parts, " ")
		}
	}
	return id
}
