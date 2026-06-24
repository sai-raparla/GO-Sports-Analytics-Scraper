package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// playerIDRe validates a baseball-reference player ID (e.g. "judgeaa01").
var playerIDRe = regexp.MustCompile(`^[a-z][a-z0-9'.-]+[0-9]{2}$`)

// ValidatePlayerID reports whether id looks like a baseball-reference player ID.
func ValidatePlayerID(id string) bool {
	return playerIDRe.MatchString(strings.ToLower(id))
}

// playerURL builds the canonical player page URL from a player ID. The first
// letter of the ID is the directory shard used by the site.
func playerURL(id string) string {
	id = strings.ToLower(id)
	return fmt.Sprintf("%s/players/%s/%s.shtml", BaseURL, id[:1], id)
}

// gameLogURL builds the season game-log URL. t is "b" (batting) or "p"
// (pitching).
func gameLogURL(id, t string, year int) string {
	id = strings.ToLower(id)
	return fmt.Sprintf("%s/players/gl.fcgi?id=%s&t=%s&year=%d", BaseURL, id, t, year)
}

// teamURL builds the canonical team season page URL.
func teamURL(id string, year int) string {
	return fmt.Sprintf("%s/teams/%s/%d.shtml", BaseURL, strings.ToUpper(id), year)
}

// rowCells returns a map of data-stat -> cell text for a single table row.
// baseball-reference tags every cell with a data-stat attribute, which is a
// far more stable key than column position.
func rowCells(row *goquery.Selection) map[string]string {
	cells := make(map[string]string)
	row.Find("th, td").Each(func(_ int, c *goquery.Selection) {
		stat, ok := c.Attr("data-stat")
		if !ok {
			return
		}
		cells[stat] = strings.TrimSpace(c.Text())
	})
	return cells
}

// isHeaderRow reports whether a row is a repeated in-body header/spacer rather
// than a data row.
func isHeaderRow(row *goquery.Selection) bool {
	class, _ := row.Attr("class")
	return strings.Contains(class, "thead") || strings.Contains(class, "spacer")
}
