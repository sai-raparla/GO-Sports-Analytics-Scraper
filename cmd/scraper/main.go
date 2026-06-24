// Command scraper is a CLI for scraping MLB player stats from
// baseball-reference.com into JSON/CSV files.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"sports-scraper/internal/models"
	"sports-scraper/internal/output"
	"sports-scraper/internal/scraper"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "scraper",
		Short: "Scrape MLB player stats from baseball-reference.com",
		Long: "scraper collects historical (season/career) and recent (game-log)\n" +
			"baseball stats from baseball-reference.com and writes them to JSON/CSV.\n\n" +
			"Please scrape responsibly: the tool is rate-limited by default to respect\n" +
			"baseball-reference's servers and terms of use.",
	}
	root.AddCommand(playerCmd(), gameLogCmd(), searchCmd())
	return root
}

func playerCmd() *cobra.Command {
	var id, name, format, outDir string
	cmd := &cobra.Command{
		Use:   "player",
		Short: "Scrape a player's bio and season/career stats (historical)",
		RunE: func(_ *cobra.Command, _ []string) error {
			resolvedID, err := resolveID(id, name)
			if err != nil {
				return err
			}

			fmt.Printf("Scraping player %s ...\n", resolvedID)
			player, err := scraper.FetchPlayer(resolvedID)
			if err != nil {
				return err
			}
			fmt.Printf("Found %s (%d batting, %d pitching seasons)\n",
				player.Name, len(player.Batting), len(player.Pitching))

			return writePlayer(format, outDir, resolvedID, player)
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "baseball-reference player id (e.g. judgeaa01)")
	cmd.Flags().StringVar(&name, "name", "", "player name to resolve to an id (alternative to --id)")
	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, or both")
	cmd.Flags().StringVar(&outDir, "out", "output", "output directory")
	return cmd
}

func gameLogCmd() *cobra.Command {
	var id, name, logType, format, outDir string
	var year int
	cmd := &cobra.Command{
		Use:   "gamelog",
		Short: "Scrape a player's game-by-game log for a season (recent/live)",
		RunE: func(_ *cobra.Command, _ []string) error {
			resolvedID, err := resolveID(id, name)
			if err != nil {
				return err
			}

			fmt.Printf("Scraping %s game log for %s %d ...\n", logType, resolvedID, year)
			logs, err := scraper.FetchGameLog(resolvedID, logType, year)
			if err != nil {
				return err
			}
			fmt.Printf("Found %d games\n", len(logs))
			if len(logs) == 0 {
				fmt.Println("(no games found - check the year and log type)")
				return nil
			}

			base := fmt.Sprintf("%s_%d_%s_gamelog", resolvedID, year, logType)
			if format == "json" || format == "both" {
				path := filepath.Join(outDir, base+".json")
				if err := output.WriteJSON(path, logs); err != nil {
					return err
				}
				fmt.Println("wrote", path)
			}
			if format == "csv" || format == "both" {
				path := filepath.Join(outDir, base+".csv")
				if err := output.WriteGameLogCSV(path, logs); err != nil {
					return err
				}
				fmt.Println("wrote", path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "baseball-reference player id (e.g. judgeaa01)")
	cmd.Flags().StringVar(&name, "name", "", "player name to resolve to an id (alternative to --id)")
	cmd.Flags().IntVar(&year, "year", time.Now().Year(), "season year")
	cmd.Flags().StringVar(&logType, "type", "batting", "log type: batting or pitching")
	cmd.Flags().StringVar(&format, "format", "json", "output format: json, csv, or both")
	cmd.Flags().StringVar(&outDir, "out", "output", "output directory")
	return cmd
}

func searchCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Find a player's id by (partial) name",
		RunE: func(_ *cobra.Command, _ []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			fmt.Printf("Searching for %q ...\n", name)
			results, err := scraper.SearchPlayers(name)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Println("no matches found")
				return nil
			}
			fmt.Printf("%-14s  %-26s  %-10s  %s\n", "ID", "NAME", "YEARS", "ACTIVE")
			for _, r := range results {
				active := ""
				if r.Active {
					active = "yes"
				}
				fmt.Printf("%-14s  %-26s  %-10s  %s\n", r.ID, r.Name, r.Years, active)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "player name to search for")
	return cmd
}

// resolveID returns id directly when provided, otherwise resolves a name to a
// single player id, erroring on ambiguity so the user can disambiguate.
func resolveID(id, name string) (string, error) {
	if id != "" {
		return strings.ToLower(id), nil
	}
	if name == "" {
		return "", fmt.Errorf("provide --id or --name")
	}
	results, err := scraper.SearchPlayers(name)
	if err != nil {
		return "", err
	}
	switch len(results) {
	case 0:
		return "", fmt.Errorf("no player found matching %q", name)
	case 1:
		fmt.Printf("Resolved %q -> %s (%s)\n", name, results[0].ID, results[0].Name)
		return results[0].ID, nil
	default:
		var b strings.Builder
		fmt.Fprintf(&b, "%q matched %d players; re-run with --id:\n", name, len(results))
		for _, r := range results {
			fmt.Fprintf(&b, "  %-14s %s (%s)\n", r.ID, r.Name, r.Years)
		}
		return "", fmt.Errorf("%s", b.String())
	}
}

func writePlayer(format, outDir, id string, player *models.Player) error {
	if format == "json" || format == "both" {
		path := filepath.Join(outDir, id+".json")
		if err := output.WriteJSON(path, player); err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}
	if format == "csv" || format == "both" {
		paths, err := output.WritePlayerCSV(outDir, id, player)
		if err != nil {
			return err
		}
		for _, p := range paths {
			fmt.Println("wrote", p)
		}
	}
	return nil
}
