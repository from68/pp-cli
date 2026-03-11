package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/from68/pp-cli/internal/format"
	"github.com/from68/pp-cli/internal/model"
	"github.com/spf13/cobra"
)

var includeRetired bool

var securitiesCmd = &cobra.Command{
	Use:   "securities",
	Short: "Work with securities",
}

var securitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List securities",
	RunE:  runSecuritiesList,
}

var securitiesShowCmd = &cobra.Command{
	Use:   "show <name-or-uuid>",
	Short: "Show security detail with price history",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecuritiesShow,
}

func init() {
	securitiesListCmd.Flags().BoolVar(&includeRetired, "retired", false, "Include retired securities")
	securitiesCmd.AddCommand(securitiesListCmd, securitiesShowCmd)
	rootCmd.AddCommand(securitiesCmd)
}

func runSecuritiesList(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()

	headers := []string{"Name", "ISIN", "Ticker", "Currency", "Price Count", "Latest Price", "Latest Date"}
	var rows [][]string
	type jsonRow struct {
		Name        string `json:"name"`
		ISIN        string `json:"isin"`
		Ticker      string `json:"ticker"`
		Currency    string `json:"currency"`
		PriceCount  int    `json:"price_count"`
		LatestPrice string `json:"latest_price"`
		LatestDate  string `json:"latest_date"`
	}
	var jsonRows []jsonRow

	for _, s := range client.Securities {
		if s.Retired && !includeRetired {
			continue
		}
		latestPrice, latestDate := secLatestPrice(s)
		row := []string{
			s.Name, s.ISIN, s.Ticker, s.Currency,
			fmt.Sprintf("%d", len(s.Prices)),
			latestPrice, latestDate,
		}
		rows = append(rows, row)
		jsonRows = append(jsonRows, jsonRow{
			Name: s.Name, ISIN: s.ISIN, Ticker: s.Ticker, Currency: s.Currency,
			PriceCount: len(s.Prices), LatestPrice: latestPrice, LatestDate: latestDate,
		})
	}

	return format.Write(w, outFmt, headers, rows, jsonRows)
}

func runSecuritiesShow(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()
	query := args[0]

	sec := findSecurity(client, query)
	if sec == nil {
		fmt.Fprintf(os.Stderr, "error: security %q not found\n", query)
		os.Exit(1)
	}

	latestPrice, latestDate := secLatestPrice(sec)

	// Print detail header.
	fmt.Fprintf(w, "Name:       %s\n", sec.Name)
	fmt.Fprintf(w, "UUID:       %s\n", sec.UUID)
	fmt.Fprintf(w, "ISIN:       %s\n", sec.ISIN)
	fmt.Fprintf(w, "Ticker:     %s\n", sec.Ticker)
	fmt.Fprintf(w, "Currency:   %s\n", sec.Currency)
	fmt.Fprintf(w, "Retired:    %v\n", sec.Retired)
	fmt.Fprintf(w, "Feed:       %s\n", sec.Feed)
	fmt.Fprintf(w, "Latest:     %s on %s\n", latestPrice, latestDate)
	fmt.Fprintf(w, "Prices:     %d\n\n", len(sec.Prices))

	// Price history table.
	if len(sec.Prices) > 0 {
		headers := []string{"Date", "Value"}
		var rows [][]string
		type jsonPrice struct {
			Date  string  `json:"date"`
			Value float64 `json:"value"`
		}
		var jsonRows []jsonPrice
		for _, p := range sec.Prices {
			v := fmt.Sprintf("%.2f", float64(p.Value)/100.0)
			rows = append(rows, []string{p.Date.Format("2006-01-02"), v})
			jsonRows = append(jsonRows, jsonPrice{Date: p.Date.Format("2006-01-02"), Value: float64(p.Value) / 100.0})
		}
		return format.Write(w, outFmt, headers, rows, jsonRows)
	}
	return nil
}

// findSecurity searches by name (case-insensitive substring) or UUID.
func findSecurity(client *model.Client, query string) *model.Security {
	lower := strings.ToLower(query)
	for _, s := range client.Securities {
		if s.UUID == query || strings.ToLower(s.Name) == lower {
			return s
		}
	}
	// Substring match.
	for _, s := range client.Securities {
		if strings.Contains(strings.ToLower(s.Name), lower) {
			return s
		}
	}
	return nil
}

// secLatestPrice returns the formatted latest price and date for a security.
func secLatestPrice(s *model.Security) (price, date string) {
	if len(s.Prices) == 0 {
		return "—", "—"
	}
	p := s.Prices[len(s.Prices)-1]
	return fmt.Sprintf("%.2f", float64(p.Value)/100.0), p.Date.Format("2006-01-02")
}
