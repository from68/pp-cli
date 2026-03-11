package commands

import (
	"fmt"
	"os"

	"github.com/from68/pp-cli/internal/model"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate portfolio file integrity",
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

type validationResult struct {
	checks   int
	errors   []string
	warnings []string
}

func (r *validationResult) check() {
	r.checks++
}

func (r *validationResult) error(msg string) {
	r.checks++
	r.errors = append(r.errors, msg)
}

func (r *validationResult) warn(msg string) {
	r.warnings = append(r.warnings, msg)
}

func runValidate(cmd *cobra.Command, args []string) error {
	client := clientFromContext(cmd)
	w := cmd.OutOrStdout()

	res := &validationResult{}

	checkUnresolvedRefs(client, res)
	checkNonNegativeAmounts(client, res)
	checkCrossEntryConsistency(client, res)

	// Print errors and warnings.
	for _, e := range res.errors {
		fmt.Fprintf(w, "ERROR: %s\n", e)
	}
	for _, wn := range res.warnings {
		fmt.Fprintf(w, "WARN:  %s\n", wn)
	}

	fmt.Fprintf(w, "\nChecks: %d, Errors: %d, Warnings: %d\n",
		res.checks, len(res.errors), len(res.warnings))

	if len(res.errors) > 0 {
		os.Exit(1)
	}
	return nil
}

// checkUnresolvedRefs reports account/portfolio transactions with unresolved security refs.
func checkUnresolvedRefs(client *model.Client, res *validationResult) {
	for _, acc := range client.Accounts {
		for i, tx := range acc.Transactions {
			if tx.SecurityRef == "" {
				continue
			}
			res.check()
			if tx.Security == nil {
				res.error(fmt.Sprintf("account %q transaction[%d]: unresolved reference %q", acc.Name, i, tx.SecurityRef))
			}
		}
	}
	for _, pf := range client.Portfolios {
		for i, tx := range pf.Transactions {
			if tx.SecurityRef == "" {
				continue
			}
			res.check()
			if tx.Security == nil {
				res.error(fmt.Sprintf("portfolio %q transaction[%d]: unresolved reference %q", pf.Name, i, tx.SecurityRef))
			}
		}
	}
}

// checkNonNegativeAmounts verifies all transaction amounts and shares are non-negative.
func checkNonNegativeAmounts(client *model.Client, res *validationResult) {
	for _, acc := range client.Accounts {
		for i, tx := range acc.Transactions {
			res.check()
			if tx.Amount < 0 {
				res.error(fmt.Sprintf("account %q transaction[%d] %s has negative amount: %d", acc.Name, i, tx.Type, tx.Amount))
			}
		}
	}
	for _, pf := range client.Portfolios {
		for i, tx := range pf.Transactions {
			res.check()
			if tx.Amount < 0 {
				res.error(fmt.Sprintf("portfolio %q transaction[%d] %s has negative amount: %d", pf.Name, i, tx.Type, tx.Amount))
			}
			res.check()
			if tx.Shares < 0 {
				res.error(fmt.Sprintf("portfolio %q transaction[%d] %s has negative shares: %d", pf.Name, i, tx.Type, int64(tx.Shares)))
			}
		}
	}
}

// checkCrossEntryConsistency verifies buy/sell cross-entry amount consistency.
// Portfolio Performance stores cross-entries with matching amounts; we check that
// portfolio transactions tagged BUY/SELL have a corresponding account transaction
// on the same date with the same amount.
func checkCrossEntryConsistency(client *model.Client, res *validationResult) {
	// Build a map of account transactions by date+amount for cross-checking.
	type key struct {
		date   string
		amount int64
	}
	accTxMap := make(map[key]bool)
	for _, acc := range client.Accounts {
		for _, tx := range acc.Transactions {
			if tx.Type == "BUY" || tx.Type == "SELL" {
				accTxMap[key{tx.Date.Format("2006-01-02"), tx.Amount}] = true
			}
		}
	}

	for _, pf := range client.Portfolios {
		for i, tx := range pf.Transactions {
			if tx.Type != "BUY" && tx.Type != "SELL" {
				continue
			}
			res.check()
			k := key{tx.Date.Format("2006-01-02"), tx.Amount}
			if !accTxMap[k] {
				res.warn(fmt.Sprintf("portfolio %q transaction[%d] %s on %s amount=%d has no matching account cross-entry",
					pf.Name, i, tx.Type, tx.Date.Format("2006-01-02"), tx.Amount))
			}
		}
	}
}
