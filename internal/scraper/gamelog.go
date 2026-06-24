package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	"sports-scraper/internal/models"
)

var gameDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

// FetchGameLog scrapes a player's game-by-game log for a season. This is the
// freshest ("live"/recent) data baseball-reference exposes for a player.
// logType is "batting" or "pitching".
func FetchGameLog(id, logType string, year int) ([]models.GameLog, error) {
	if !ValidatePlayerID(id) {
		return nil, fmt.Errorf("invalid player id %q", id)
	}

	var t, tableSel string
	switch strings.ToLower(logType) {
	case "batting", "b":
		t, tableSel, logType = "b", "#players_standard_batting", "batting"
	case "pitching", "p":
		t, tableSel, logType = "p", "#players_standard_pitching", "pitching"
	default:
		return nil, fmt.Errorf("invalid log type %q (use batting or pitching)", logType)
	}

	c := newCollector()
	url := gameLogURL(id, t, year)
	var logs []models.GameLog

	c.OnHTML(tableSel, func(e *colly.HTMLElement) {
		e.DOM.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
			if isHeaderRow(row) {
				return
			}
			cells := rowCells(row)
			date := cells["date"]
			if !gameDateRe.MatchString(date) {
				return // skip summary / non-game rows
			}
			logs = append(logs, gameLogFromCells(id, logType, year, cells))
		})
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
	return logs, nil
}

// metaCols are the descriptive columns lifted into dedicated GameLog fields;
// everything else becomes a stat in the Stats map.
var metaCols = map[string]bool{
	"date":           true,
	"team_name_abbr": true,
	"opp_name_abbr":  true,
	"ranker":         true,
}

func gameLogFromCells(id, logType string, year int, c map[string]string) models.GameLog {
	stats := make(map[string]string)
	for k, v := range c {
		if metaCols[k] || v == "" {
			continue
		}
		stats[k] = v
	}
	return models.GameLog{
		PlayerID: strings.ToLower(id),
		Year:     year,
		Type:     logType,
		Date:     c["date"],
		Team:     c["team_name_abbr"],
		Opponent: c["opp_name_abbr"],
		Stats:    stats,
	}
}
