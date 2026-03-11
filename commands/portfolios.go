package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/from68/pp-cli/internal/format"
	"github.com/from68/pp-cli/internal/model"
	"github.com/spf13/cobra"
)

var portfoliosCmd = &cobra.Command{
	Use:   "portfolios",
	Short: "Work with portfolios",
}

var portfoliosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List portfolios",
	RunE:  runPortfoliosList,
}

var (
	pfTxFrom  string
	pfTxTo    string
	pfTxTypes string
)

var portfoliosTxCmd = &cobra.Command{
	Use:   "transactions <portfolio-name>",
	Short: "List transactions for a portfolio",
	Args:  cobra.ExactArgs(1),
	RunE:  runPortfoliosTransactions,
}

func init() {
	portfoliosTxCmd.Flags().StringVar(&pfTxFrom, "from", "", "Start date (ISO 8601, inclusive)")
	portfoliosTxCmd.Flags().StringVar(&pfTxTo, "to", "", "End date (ISO 8601, inclusive)")
	portfoliosTxCmd.Flags().StringVar(&pfTxTypes, "type", "", "Filter by type(s), comma-separated")
	portfoliosCmd.AddCommand(portfoliosListCmd, portfoliosTxCmd)
	rootCmd.AddCommand(portfoliosCmd)
}

func runPortfoliosList(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()

	headers := []string{"Name", "Reference Account"}
	var rows [][]string
	type jsonRow struct {
		Name             string `json:"name"`
		ReferenceAccount string `json:"reference_account"`
	}
	var jsonRows []jsonRow

	for _, pf := range client.Portfolios {
		refAcc := "—"
		if pf.ReferenceAccount != nil {
			refAcc = pf.ReferenceAccount.Name
		}
		rows = append(rows, []string{pf.Name, refAcc})
		jsonRows = append(jsonRows, jsonRow{Name: pf.Name, ReferenceAccount: refAcc})
	}

	return format.Write(w, outFmt, headers, rows, jsonRows)
}

func runPortfoliosTransactions(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()
	name := args[0]

	pf := findPortfolio(client, name)
	if pf == nil {
		fmt.Fprintf(os.Stderr, "error: portfolio %q not found\n", name)
		os.Exit(1)
	}

	from, to, err := parseDateRange(pfTxFrom, pfTxTo)
	if err != nil {
		return err
	}

	allowedTypes := parseTypes(pfTxTypes)

	headers := []string{"Date", "Type", "Shares", "Amount", "Currency", "Security", "Note"}
	var rows [][]string
	type jsonRow struct {
		Date     string  `json:"date"`
		Type     string  `json:"type"`
		Shares   string  `json:"shares"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
		Security string  `json:"security"`
		Note     string  `json:"note"`
	}
	var jsonRows []jsonRow

	for _, tx := range pf.Transactions {
		if !inDateRange(tx.Date, from, to) {
			continue
		}
		if len(allowedTypes) > 0 && !allowedTypes[strings.ToUpper(tx.Type)] {
			continue
		}

		secName := "—"
		if tx.Security != nil {
			secName = tx.Security.Name
		}
		sharesStr := fmt.Sprintf("%.8f", tx.Shares.Value())
		amtStr := fmt.Sprintf("%.2f", float64(tx.Amount)/100.0)
		rows = append(rows, []string{
			tx.Date.Format("2006-01-02"),
			tx.Type,
			sharesStr,
			amtStr,
			tx.Currency,
			secName,
			tx.Note,
		})
		jsonRows = append(jsonRows, jsonRow{
			Date: tx.Date.Format("2006-01-02"), Type: tx.Type,
			Shares: sharesStr, Amount: float64(tx.Amount) / 100.0,
			Currency: tx.Currency, Security: secName, Note: tx.Note,
		})
	}

	return format.Write(w, outFmt, headers, rows, jsonRows)
}

func findPortfolio(client *model.Client, name string) *model.Portfolio {
	lower := strings.ToLower(name)
	for _, p := range client.Portfolios {
		if strings.ToLower(p.Name) == lower || p.UUID == name {
			return p
		}
	}
	return nil
}

// parseTypes splits a comma-separated type list into a set.
func parseTypes(s string) map[string]bool {
	if s == "" {
		return nil
	}
	m := make(map[string]bool)
	for _, t := range strings.Split(s, ",") {
		m[strings.ToUpper(strings.TrimSpace(t))] = true
	}
	return m
}
