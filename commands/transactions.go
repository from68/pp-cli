package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/from68/pp-cli/internal/format"
	"github.com/from68/pp-cli/internal/model"
	"github.com/spf13/cobra"
)

var (
	txFrom     string
	txTo       string
	txTypes    string
	txSecurity string
)

var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Show all transactions merged and sorted by date",
	RunE:  runTransactions,
}

func init() {
	transactionsCmd.Flags().StringVar(&txFrom, "from", "", "Start date (ISO 8601, inclusive)")
	transactionsCmd.Flags().StringVar(&txTo, "to", "", "End date (ISO 8601, inclusive)")
	transactionsCmd.Flags().StringVar(&txTypes, "type", "", "Filter by type(s), comma-separated")
	transactionsCmd.Flags().StringVar(&txSecurity, "security", "", "Filter by security name or ISIN")
	rootCmd.AddCommand(transactionsCmd)
}

type unifiedTx struct {
	Date     time.Time
	Source   string
	Type     string
	Amount   int64
	Currency string
	Shares   model.Shares
	Security *model.Security
	Note     string
}

func runTransactions(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()

	from, to, err := parseDateRange(txFrom, txTo)
	if err != nil {
		return err
	}
	allowedTypes := parseTypes(txTypes)

	var all []unifiedTx

	for _, acc := range client.Accounts {
		for _, tx := range acc.Transactions {
			all = append(all, unifiedTx{
				Date:     tx.Date,
				Source:   acc.Name,
				Type:     tx.Type,
				Amount:   tx.Amount,
				Currency: tx.Currency,
				Security: tx.Security,
				Note:     tx.Note,
			})
		}
	}
	for _, pf := range client.Portfolios {
		for _, tx := range pf.Transactions {
			all = append(all, unifiedTx{
				Date:     tx.Date,
				Source:   pf.Name,
				Type:     tx.Type,
				Amount:   tx.Amount,
				Currency: tx.Currency,
				Shares:   tx.Shares,
				Security: tx.Security,
				Note:     tx.Note,
			})
		}
	}

	// Sort by date ascending.
	sort.Slice(all, func(i, j int) bool {
		return all[i].Date.Before(all[j].Date)
	})

	headers := []string{"Date", "Source", "Type", "Amount", "Currency", "Shares", "Security", "Note"}
	var rows [][]string
	type jsonRow struct {
		Date     string  `json:"date"`
		Source   string  `json:"source"`
		Type     string  `json:"type"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
		Shares   string  `json:"shares"`
		Security string  `json:"security"`
		Note     string  `json:"note"`
	}
	var jsonRows []jsonRow

	for _, tx := range all {
		if !inDateRange(tx.Date, from, to) {
			continue
		}
		if len(allowedTypes) > 0 && !allowedTypes[strings.ToUpper(tx.Type)] {
			continue
		}
		if txSecurity != "" && !matchesSecurity(tx.Security, txSecurity) {
			continue
		}

		secName := "—"
		if tx.Security != nil {
			secName = tx.Security.Name
		}
		sharesStr := "—"
		if tx.Shares != 0 {
			sharesStr = fmt.Sprintf("%.8f", tx.Shares.Value())
		}
		amtStr := fmt.Sprintf("%.2f", float64(tx.Amount)/100.0)

		rows = append(rows, []string{
			tx.Date.Format("2006-01-02"),
			tx.Source,
			tx.Type,
			amtStr,
			tx.Currency,
			sharesStr,
			secName,
			tx.Note,
		})
		jsonRows = append(jsonRows, jsonRow{
			Date: tx.Date.Format("2006-01-02"), Source: tx.Source, Type: tx.Type,
			Amount: float64(tx.Amount) / 100.0, Currency: tx.Currency,
			Shares: sharesStr, Security: secName, Note: tx.Note,
		})
	}

	return format.Write(w, outFmt, headers, rows, jsonRows)
}

func matchesSecurity(sec *model.Security, query string) bool {
	if sec == nil {
		return false
	}
	lower := strings.ToLower(query)
	return strings.ToLower(sec.Name) == lower ||
		strings.EqualFold(sec.ISIN, query) ||
		strings.Contains(strings.ToLower(sec.Name), lower)
}
