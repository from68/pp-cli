package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/from68/pp-cli/internal/format"
	"github.com/from68/pp-cli/internal/model"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Work with accounts",
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accounts with computed balance",
	RunE:  runAccountsList,
}

var (
	accTxFrom string
	accTxTo   string
	accTxType string
)

var accountsTxCmd = &cobra.Command{
	Use:   "transactions <account-name>",
	Short: "List transactions for an account",
	Args:  cobra.ExactArgs(1),
	RunE:  runAccountsTransactions,
}

func init() {
	accountsTxCmd.Flags().StringVar(&accTxFrom, "from", "", "Start date (ISO 8601, inclusive)")
	accountsTxCmd.Flags().StringVar(&accTxTo, "to", "", "End date (ISO 8601, inclusive)")
	accountsTxCmd.Flags().StringVar(&accTxType, "type", "", "Filter by transaction type")
	accountsCmd.AddCommand(accountsListCmd, accountsTxCmd)
	rootCmd.AddCommand(accountsCmd)
}

func runAccountsList(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()

	headers := []string{"Name", "Currency", "Balance", "Note"}
	var rows [][]string
	type jsonRow struct {
		Name     string  `json:"name"`
		Currency string  `json:"currency"`
		Balance  float64 `json:"balance"`
		Note     string  `json:"note"`
	}
	var jsonRows []jsonRow

	for _, acc := range client.Accounts {
		bal := computeBalance(acc)
		balStr := fmt.Sprintf("%.2f", float64(bal)/100.0)
		rows = append(rows, []string{acc.Name, acc.Currency, balStr, acc.Note})
		jsonRows = append(jsonRows, jsonRow{Name: acc.Name, Currency: acc.Currency, Balance: float64(bal) / 100.0, Note: acc.Note})
	}

	return format.Write(w, outFmt, headers, rows, jsonRows)
}

// computeBalance sums all cash-positive and cash-negative transactions.
func computeBalance(acc *model.Account) int64 {
	var bal int64
	for _, tx := range acc.Transactions {
		switch tx.Type {
		case "DEPOSIT", "TRANSFER_IN", "SELL", "INTEREST", "DIVIDENDS":
			bal += tx.Amount
		case "REMOVAL", "TRANSFER_OUT", "BUY", "INTEREST_CHARGE", "FEES":
			bal -= tx.Amount
		}
	}
	return bal
}

func runAccountsTransactions(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	outFmt := formatFromContext(cmd)
	w := cmd.OutOrStdout()
	name := args[0]

	acc := findAccount(client, name)
	if acc == nil {
		fmt.Fprintf(os.Stderr, "error: account %q not found\n", name)
		os.Exit(1)
	}

	from, to, err := parseDateRange(accTxFrom, accTxTo)
	if err != nil {
		return err
	}

	headers := []string{"Date", "Type", "Amount", "Currency", "Shares", "Security", "Note"}
	var rows [][]string
	type jsonRow struct {
		Date     string  `json:"date"`
		Type     string  `json:"type"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
		Shares   string  `json:"shares"`
		Security string  `json:"security"`
		Note     string  `json:"note"`
	}
	var jsonRows []jsonRow

	for _, tx := range acc.Transactions {
		if !inDateRange(tx.Date, from, to) {
			continue
		}
		if accTxType != "" && !strings.EqualFold(tx.Type, accTxType) {
			continue
		}

		secName := "—"
		if tx.Security != nil {
			secName = tx.Security.Name
		}
		sharesStr := "—"
		// Account transactions may have share units.
		for _, u := range tx.Units {
			if u.Shares != 0 {
				sharesStr = fmt.Sprintf("%.8f", u.Shares.Value())
				break
			}
		}
		amtStr := fmt.Sprintf("%.2f", float64(tx.Amount)/100.0)
		row := []string{
			tx.Date.Format("2006-01-02"),
			tx.Type,
			amtStr,
			tx.Currency,
			sharesStr,
			secName,
			tx.Note,
		}
		rows = append(rows, row)
		jsonRows = append(jsonRows, jsonRow{
			Date: tx.Date.Format("2006-01-02"), Type: tx.Type,
			Amount: float64(tx.Amount) / 100.0, Currency: tx.Currency,
			Shares: sharesStr, Security: secName, Note: tx.Note,
		})
	}

	return format.Write(w, outFmt, headers, rows, jsonRows)
}

func findAccount(client *model.Client, name string) *model.Account {
	lower := strings.ToLower(name)
	for _, a := range client.Accounts {
		if strings.ToLower(a.Name) == lower || a.UUID == name {
			return a
		}
	}
	return nil
}

func parseDateRange(fromStr, toStr string) (from, to time.Time, err error) {
	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --from date: %w", err)
		}
	}
	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --to date: %w", err)
		}
		// Make to inclusive by setting to end-of-day.
		to = to.Add(24*time.Hour - time.Nanosecond)
	}
	return from, to, nil
}

func inDateRange(t, from, to time.Time) bool {
	if !from.IsZero() && t.Before(from) {
		return false
	}
	if !to.IsZero() && t.After(to) {
		return false
	}
	return true
}
