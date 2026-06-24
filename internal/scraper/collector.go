// Package scraper contains the colly-based scraping logic for
// baseball-reference.com.
package scraper

import (
	"bytes"
	"log"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	// BaseURL is the root of the site we scrape.
	BaseURL = "https://www.baseball-reference.com"

	// userAgent identifies the scraper. Using a real, descriptive UA is more
	// polite than spoofing a browser and helps the site contact us if needed.
	userAgent = "GO-Sports-Analytics-Scraper/0.1 (+https://github.com/; educational use)"

	// maxRetries bounds how many times a single request is retried on error
	// (notably HTTP 429 rate limiting).
	maxRetries = 4
)

// newCollector builds a colly.Collector configured for baseball-reference:
//   - a descriptive User-Agent
//   - a polite LimitRule (single connection, delay + jitter) to stay under the
//     site's ~20 req/min ceiling
//   - an OnResponse hook that strips HTML comment markers so that the many
//     stat tables baseball-reference hides inside <!-- ... --> become real DOM
//     nodes visible to OnHTML selectors
//   - an OnError hook that retries with exponential backoff, which is essential
//     for surviving 429 responses
func newCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains("www.baseball-reference.com", "baseball-reference.com"),
	)

	// Be a good citizen: one request at a time with a delay and some jitter.
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*baseball-reference.com*",
		Parallelism: 1,
		Delay:       3 * time.Second,
		RandomDelay: 2 * time.Second,
	}); err != nil {
		log.Printf("warning: failed to set rate limit: %v", err)
	}

	// baseball-reference wraps most stat tables in HTML comments to deter
	// scrapers. colly's OnHTML parses the response body, and the Go HTML parser
	// keeps commented markup as inert comment nodes that selectors cannot
	// match. Removing the comment delimiters promotes that markup to live DOM.
	c.OnResponse(func(r *colly.Response) {
		r.Body = stripHTMLComments(r.Body)
	})

	c.OnError(func(r *colly.Response, err error) {
		retries := getRetryCount(r)
		if retries >= maxRetries {
			log.Printf("giving up on %s after %d retries: %v", r.Request.URL, retries, err)
			return
		}

		// Exponential backoff: 2s, 4s, 8s, ... 429s in particular need a real
		// pause before we try again.
		backoff := time.Duration(1<<retries) * 2 * time.Second
		log.Printf("request to %s failed (%v); retry %d/%d after %s",
			r.Request.URL, err, retries+1, maxRetries, backoff)
		time.Sleep(backoff)

		setRetryCount(r, retries+1)
		if rerr := r.Request.Retry(); rerr != nil {
			log.Printf("retry scheduling failed for %s: %v", r.Request.URL, rerr)
		}
	})

	return c
}

// stripHTMLComments removes HTML comment delimiters (<!-- and -->) from the
// body so that commented-out tables become parseable markup.
func stripHTMLComments(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte("<!--"), []byte(""))
	body = bytes.ReplaceAll(body, []byte("-->"), []byte(""))
	return body
}

// retryCountKey is the colly request-context key used to track per-request
// retry attempts.
const retryCountKey = "retry_count"

func getRetryCount(r *colly.Response) int {
	if r == nil || r.Request == nil || r.Request.Ctx == nil {
		return 0
	}
	v := r.Request.Ctx.Get(retryCountKey)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func setRetryCount(r *colly.Response, n int) {
	if r == nil || r.Request == nil || r.Request.Ctx == nil {
		return
	}
	r.Request.Ctx.Put(retryCountKey, strconv.Itoa(n))
}
