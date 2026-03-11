package commands

import (
	"fmt"
	"time"

	"github.com/from68/pp-cli/internal/model"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display portfolio file summary",
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)

	earliest, latest, hasTransactions := transactionDateRange(client)

	fmt.Fprintf(cmd.OutOrStdout(), "Version:     %d\n", client.Version)
	fmt.Fprintf(cmd.OutOrStdout(), "Currency:    %s\n", client.BaseCurrency)
	fmt.Fprintf(cmd.OutOrStdout(), "Securities:  %d\n", len(client.Securities))
	fmt.Fprintf(cmd.OutOrStdout(), "Accounts:    %d\n", len(client.Accounts))
	fmt.Fprintf(cmd.OutOrStdout(), "Portfolios:  %d\n", len(client.Portfolios))

	if hasTransactions {
		fmt.Fprintf(cmd.OutOrStdout(), "Date Range:  %s — %s\n",
			earliest.Format("2006-01-02"),
			latest.Format("2006-01-02"),
		)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Date Range:  —\n")
	}

	return nil
}

// transactionDateRange returns the earliest and latest transaction dates across all
// accounts and portfolios, and whether any transactions exist.
func transactionDateRange(client *model.Client) (earliest, latest time.Time, hasAny bool) {
	update := func(t time.Time) {
		if t.IsZero() {
			return
		}
		if !hasAny || t.Before(earliest) {
			earliest = t
		}
		if !hasAny || t.After(latest) {
			latest = t
		}
		hasAny = true
	}

	for _, acc := range client.Accounts {
		for _, tx := range acc.Transactions {
			update(tx.Date)
		}
	}
	for _, pf := range client.Portfolios {
		for _, tx := range pf.Transactions {
			update(tx.Date)
		}
	}
	return
}
