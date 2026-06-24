package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	"sports-scraper/internal/models"
)

// FetchTeam scrapes a team's batting and pitching totals for one season.
func FetchTeam(id string, year int) (*models.Team, error) {
	if !ValidateTeamID(id) {
		return nil, fmt.Errorf("invalid team id %q (expected something like NYY)", id)
	}

	c := newCollector()
	id = strings.ToUpper(id)
	url := teamURL(id, year)
	team := &models.Team{ID: id, Year: year, SourceURL: url}

	c.OnHTML("#meta", func(e *colly.HTMLElement) {
		team.Name = parseTeamName(e.DOM, year)
	})

	c.OnHTML("#players_standard_batting, #team_batting", func(e *colly.HTMLElement) {
		team.BattingTotals = parseTeamTotals(e.DOM)
	})

	c.OnHTML("#players_standard_pitching, #team_pitching", func(e *colly.HTMLElement) {
		team.PitchingTotals = parseTeamTotals(e.DOM)
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
	if team.Name == "" && len(team.BattingTotals) == 0 && len(team.PitchingTotals) == 0 {
		return nil, fmt.Errorf("no data found for team %q in %d (does the page exist?)", id, year)
	}
	return team, nil
}

func parseTeamName(meta *goquery.Selection, year int) string {
	name := strings.Join(strings.Fields(meta.Find("h1").First().Text()), " ")
	if name == "" {
		return ""
	}
	prefix := fmt.Sprintf("%d ", year)
	name = strings.TrimPrefix(name, prefix)
	name = strings.TrimSuffix(name, " Statistics")
	return strings.TrimSpace(name)
}

func parseTeamTotals(table *goquery.Selection) map[string]string {
	var totals map[string]string
	table.Find("tfoot tr").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		if isHeaderRow(row) {
			return true
		}
		cells := rowCells(row)
		if teamPlayerName(cells) == "" {
			return true
		}
		totals = teamTotalsFromCells(cells)
		return false
	})
	return totals
}

var teamMetaCols = map[string]bool{
	"ranker":       true,
	"player":       true,
	"name_display": true,
	"age":          true,
}

func teamTotalsFromCells(c map[string]string) map[string]string {
	stats := make(map[string]string)
	for k, v := range c {
		if teamMetaCols[k] || v == "" {
			continue
		}
		stats[k] = v
	}
	return stats
}

func teamPlayerName(c map[string]string) string {
	if c["name_display"] != "" {
		return c["name_display"]
	}
	return c["player"]
}
