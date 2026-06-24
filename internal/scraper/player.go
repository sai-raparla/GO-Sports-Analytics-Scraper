package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	"sports-scraper/internal/models"
)

var careerLabelRe = regexp.MustCompile(`^\d+\s+Yrs?$`)

// FetchPlayer scrapes a player's bio plus their standard batting and/or
// pitching season tables (historical stats).
func FetchPlayer(id string) (*models.Player, error) {
	if !ValidatePlayerID(id) {
		return nil, fmt.Errorf("invalid player id %q (expected something like judgeaa01)", id)
	}

	c := newCollector()
	url := playerURL(id)
	player := &models.Player{ID: strings.ToLower(id), SourceURL: url}

	c.OnHTML("#meta", func(e *colly.HTMLElement) {
		parseBio(e.DOM, player)
	})

	c.OnHTML("#players_standard_batting", func(e *colly.HTMLElement) {
		player.Batting = parseBattingTable(e.DOM)
	})

	c.OnHTML("#players_standard_pitching", func(e *colly.HTMLElement) {
		player.Pitching = parsePitchingTable(e.DOM)
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
	if player.Name == "" && len(player.Batting) == 0 && len(player.Pitching) == 0 {
		return nil, fmt.Errorf("no data found for player %q (does the page exist?)", id)
	}
	return player, nil
}

// parseBio extracts best-effort biographical fields from the #meta block.
func parseBio(meta *goquery.Selection, p *models.Player) {
	p.Name = strings.TrimSpace(meta.Find("h1").First().Text())

	meta.Find("p").Each(func(_ int, s *goquery.Selection) {
		text := strings.Join(strings.Fields(s.Text()), " ")
		switch {
		case strings.HasPrefix(text, "Position:"):
			p.Position = cleanLabel(text, "Position:")
			// The same paragraph often carries Bats/Throws on the site.
			if m := regexp.MustCompile(`Bats:\s*([A-Za-z]+)`).FindStringSubmatch(text); m != nil {
				p.Bats = m[1]
			}
			if m := regexp.MustCompile(`Throws:\s*([A-Za-z]+)`).FindStringSubmatch(text); m != nil {
				p.Throws = m[1]
			}
		case strings.HasPrefix(text, "Bats:"):
			if m := regexp.MustCompile(`Bats:\s*([A-Za-z]+)`).FindStringSubmatch(text); m != nil {
				p.Bats = m[1]
			}
			if m := regexp.MustCompile(`Throws:\s*([A-Za-z]+)`).FindStringSubmatch(text); m != nil {
				p.Throws = m[1]
			}
		case strings.HasPrefix(text, "Team:"):
			p.Team = cleanLabel(text, "Team:")
		case strings.HasPrefix(text, "Born:"):
			p.Born = cleanLabel(text, "Born:")
		default:
			if p.Height == "" {
				if m := regexp.MustCompile(`(\d+-\d+),?\s*(\d+lb)`).FindStringSubmatch(text); m != nil {
					p.Height = m[1]
					p.Weight = m[2]
				}
			}
		}
	})
}

func cleanLabel(text, label string) string {
	return strings.TrimSpace(strings.TrimPrefix(text, label))
}

// parseBattingTable parses both per-season rows (tbody) and the career total
// row (tfoot).
func parseBattingTable(table *goquery.Selection) []models.SeasonBatting {
	var out []models.SeasonBatting

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		if isHeaderRow(row) {
			return
		}
		cells := rowCells(row)
		if cells["year_id"] == "" {
			return
		}
		out = append(out, battingFromCells(cells["year_id"], cells))
	})

	table.Find("tfoot tr").Each(func(_ int, row *goquery.Selection) {
		cells := rowCells(row)
		label := firstCellText(row)
		if careerLabelRe.MatchString(label) {
			b := battingFromCells("Career", cells)
			out = append(out, b)
		}
	})

	return out
}

func battingFromCells(season string, c map[string]string) models.SeasonBatting {
	return models.SeasonBatting{
		Season:  season,
		Age:     c["age"],
		Team:    c["team_name_abbr"],
		League:  c["comp_name_abbr"],
		WAR:     c["b_war"],
		G:       c["b_games"],
		PA:      c["b_pa"],
		AB:      c["b_ab"],
		R:       c["b_r"],
		H:       c["b_h"],
		Doubles: c["b_doubles"],
		Triples: c["b_triples"],
		HR:      c["b_hr"],
		RBI:     c["b_rbi"],
		SB:      c["b_sb"],
		CS:      c["b_cs"],
		BB:      c["b_bb"],
		SO:      c["b_so"],
		BA:      c["b_batting_avg"],
		OBP:     c["b_onbase_perc"],
		SLG:     c["b_slugging_perc"],
		OPS:     c["b_onbase_plus_slugging"],
	}
}

// parsePitchingTable parses per-season rows and the career total row.
func parsePitchingTable(table *goquery.Selection) []models.SeasonPitching {
	var out []models.SeasonPitching

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		if isHeaderRow(row) {
			return
		}
		cells := rowCells(row)
		if cells["year_id"] == "" {
			return
		}
		out = append(out, pitchingFromCells(cells["year_id"], cells))
	})

	table.Find("tfoot tr").Each(func(_ int, row *goquery.Selection) {
		cells := rowCells(row)
		if careerLabelRe.MatchString(firstCellText(row)) {
			out = append(out, pitchingFromCells("Career", cells))
		}
	})

	return out
}

func pitchingFromCells(season string, c map[string]string) models.SeasonPitching {
	return models.SeasonPitching{
		Season: season,
		Age:    c["age"],
		Team:   c["team_name_abbr"],
		League: c["comp_name_abbr"],
		WAR:    c["p_war"],
		W:      c["p_w"],
		L:      c["p_l"],
		ERA:    c["p_earned_run_avg"],
		G:      c["p_g"],
		GS:     c["p_gs"],
		SV:     c["p_sv"],
		IP:     c["p_ip"],
		H:      c["p_h"],
		R:      c["p_r"],
		ER:     c["p_er"],
		BB:     c["p_bb"],
		SO:     c["p_so"],
		WHIP:   c["p_whip"],
	}
}

func firstCellText(row *goquery.Selection) string {
	return strings.TrimSpace(row.Find("th, td").First().Text())
}
