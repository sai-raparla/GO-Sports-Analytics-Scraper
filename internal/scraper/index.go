package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

// SearchResult is a single match from the player index.
type SearchResult struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Years  string `json:"years,omitempty"`
	Active bool   `json:"active"`
}

var (
	hrefIDRe = regexp.MustCompile(`/players/[a-z]/([a-z0-9'.-]+)\.shtml$`)
	yearsRe  = regexp.MustCompile(`\(([\d\-]+)\)`)
)

// SearchPlayers resolves a (partial) player name to candidate players by
// scraping the letter index page for the first letter of the player's last
// name. nameQuery is matched case-insensitively as a substring.
func SearchPlayers(nameQuery string) ([]SearchResult, error) {
	nameQuery = strings.TrimSpace(nameQuery)
	if nameQuery == "" {
		return nil, fmt.Errorf("empty search query")
	}

	letter, err := lastNameInitial(nameQuery)
	if err != nil {
		return nil, err
	}

	c := newCollector()
	url := fmt.Sprintf("%s/players/%s/", BaseURL, letter)
	needle := strings.ToLower(nameQuery)

	// The index links each player several times (nav, appearances, the real
	// entry). Dedupe by ID, preferring whichever occurrence carries the
	// year-range / active info.
	byID := map[string]*SearchResult{}
	var order []string

	c.OnHTML(fmt.Sprintf(`a[href^="/players/%s/"]`, letter), func(e *colly.HTMLElement) {
		href := e.Attr("href")
		m := hrefIDRe.FindStringSubmatch(href)
		if m == nil {
			return
		}
		name := strings.TrimSpace(e.Text)
		if name == "" || !strings.Contains(strings.ToLower(name), needle) {
			return
		}

		// The surrounding <p> holds the active-year range; active players are
		// rendered in bold (<b>).
		parent := e.DOM.Parent()
		years := ""
		if ym := yearsRe.FindStringSubmatch(parent.Text()); ym != nil {
			years = ym[1]
		}
		active := parent.Find("b").Length() > 0 || e.DOM.Closest("b").Length() > 0

		id := m[1]
		if existing, ok := byID[id]; ok {
			if existing.Years == "" && years != "" {
				existing.Years = years
			}
			existing.Active = existing.Active || active
			return
		}
		byID[id] = &SearchResult{ID: id, Name: name, Years: years, Active: active}
		order = append(order, id)
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

	results := make([]SearchResult, 0, len(order))
	for _, id := range order {
		results = append(results, *byID[id])
	}
	return results, nil
}

// lastNameInitial returns the first letter of the last token of a name, which
// is the directory shard baseball-reference uses for its index pages.
func lastNameInitial(name string) (string, error) {
	fields := strings.Fields(strings.ToLower(name))
	if len(fields) == 0 {
		return "", fmt.Errorf("could not parse a name from %q", name)
	}
	last := fields[len(fields)-1]
	for _, r := range last {
		if r >= 'a' && r <= 'z' {
			return string(r), nil
		}
	}
	return "", fmt.Errorf("name %q has no usable last-name initial", name)
}
