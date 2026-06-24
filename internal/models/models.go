// Package models defines the data structures produced by the scraper.
package models

// Player is the top-level record for a single baseball player, combining
// biographical data with their historical season-by-season statistics.
type Player struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position,omitempty"`
	Bats     string `json:"bats,omitempty"`
	Throws   string `json:"throws,omitempty"`
	Height   string `json:"height,omitempty"`
	Weight   string `json:"weight,omitempty"`
	Born     string `json:"born,omitempty"`
	Team     string `json:"team,omitempty"`
	SourceURL string `json:"source_url"`

	Batting  []SeasonBatting  `json:"batting,omitempty"`
	Pitching []SeasonPitching `json:"pitching,omitempty"`
}

// SeasonBatting is one row of a player's standard batting table. The "Season"
// field is "Career" for the totals row.
type SeasonBatting struct {
	Season string `json:"season"`
	Age    string `json:"age,omitempty"`
	Team   string `json:"team,omitempty"`
	League string `json:"league,omitempty"`
	WAR    string `json:"war,omitempty"`
	G      string `json:"g,omitempty"`
	PA     string `json:"pa,omitempty"`
	AB     string `json:"ab,omitempty"`
	R      string `json:"r,omitempty"`
	H      string `json:"h,omitempty"`
	Doubles string `json:"2b,omitempty"`
	Triples string `json:"3b,omitempty"`
	HR     string `json:"hr,omitempty"`
	RBI    string `json:"rbi,omitempty"`
	SB     string `json:"sb,omitempty"`
	CS     string `json:"cs,omitempty"`
	BB     string `json:"bb,omitempty"`
	SO     string `json:"so,omitempty"`
	BA     string `json:"ba,omitempty"`
	OBP    string `json:"obp,omitempty"`
	SLG    string `json:"slg,omitempty"`
	OPS    string `json:"ops,omitempty"`
}

// SeasonPitching is one row of a player's standard pitching table. The "Season"
// field is "Career" for the totals row.
type SeasonPitching struct {
	Season string `json:"season"`
	Age    string `json:"age,omitempty"`
	Team   string `json:"team,omitempty"`
	League string `json:"league,omitempty"`
	WAR    string `json:"war,omitempty"`
	W      string `json:"w,omitempty"`
	L      string `json:"l,omitempty"`
	ERA    string `json:"era,omitempty"`
	G      string `json:"g,omitempty"`
	GS     string `json:"gs,omitempty"`
	SV     string `json:"sv,omitempty"`
	IP     string `json:"ip,omitempty"`
	H      string `json:"h,omitempty"`
	R      string `json:"r,omitempty"`
	ER     string `json:"er,omitempty"`
	BB     string `json:"bb,omitempty"`
	SO     string `json:"so,omitempty"`
	WHIP   string `json:"whip,omitempty"`
}

// GameLog is one game-by-game line from a player's season game log. It is a
// flexible key/value map because batting and pitching logs have different
// columns; the stat-code -> value pairs are preserved as scraped.
type GameLog struct {
	PlayerID string            `json:"player_id"`
	Year     int               `json:"year"`
	Type     string            `json:"type"` // "batting" or "pitching"
	Date     string            `json:"date,omitempty"`
	Team     string            `json:"team,omitempty"`
	Opponent string            `json:"opponent,omitempty"`
	Stats    map[string]string `json:"stats"`
}
