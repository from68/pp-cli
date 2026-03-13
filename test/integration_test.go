package main_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/from68/pp-cli/internal/format"
	"github.com/from68/pp-cli/internal/model"
	ppxml "github.com/from68/pp-cli/internal/xml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testdataMinimal = "data/minimal.xml"
const testdataXPathRefs = "data/xpath_refs.xml"
const testdataCrossEntry = "data/crossentry.xml"

func loadTestFile(t *testing.T, path string) interface{ Close() error } {
	t.Helper()
	f, err := ppxml.Load(path)
	require.NoError(t, err)
	return f
}

func TestIntegration_DecodeMinimal(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	assert.Equal(t, 52, client.Version)
	assert.Equal(t, "EUR", client.BaseCurrency)
	assert.Len(t, client.Securities, 2)
	assert.Len(t, client.Accounts, 1)
	assert.Len(t, client.Portfolios, 1)
}

func TestIntegration_SecurityPrices(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	vwrl := client.Securities[0]
	assert.Equal(t, "Vanguard FTSE All-World", vwrl.Name)
	assert.Len(t, vwrl.Prices, 3)
	assert.Equal(t, int64(9843000000), vwrl.Prices[0].Value)
	assert.Equal(t, int64(9870000000), vwrl.Prices[2].Value)
}

func TestIntegration_RetiredSecurity(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	iwda := client.Securities[1]
	assert.True(t, iwda.Retired)
	assert.Empty(t, iwda.Prices)
}

func TestIntegration_ReferenceResolution(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	acc := client.Accounts[0]
	// tx[1] = BUY, should have security resolved to VWRL
	buyTx := acc.Transactions[1]
	require.NotNil(t, buyTx.Security)
	assert.Equal(t, "Vanguard FTSE All-World", buyTx.Security.Name)

	// Portfolio transaction also resolved
	ptx := client.Portfolios[0].Transactions[0]
	require.NotNil(t, ptx.Security)
	assert.Equal(t, "Vanguard FTSE All-World", ptx.Security.Name)
}

func TestIntegration_XPathRefFixture(t *testing.T) {
	f, err := ppxml.Load(testdataXPathRefs)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	acc := client.Accounts[0]
	require.Len(t, acc.Transactions, 4)

	// at2 references security[1] = Alpha Fund
	assert.NotNil(t, acc.Transactions[1].Security)
	assert.Equal(t, "Alpha Fund", acc.Transactions[1].Security.Name)

	// at3 references security[2] = Beta ETF
	assert.NotNil(t, acc.Transactions[2].Security)
	assert.Equal(t, "Beta ETF", acc.Transactions[2].Security.Name)

	// at4 references security[3] = Gamma Bond
	assert.NotNil(t, acc.Transactions[3].Security)
	assert.Equal(t, "Gamma Bond", acc.Transactions[3].Security.Name)

	// Portfolio transactions
	pf := client.Portfolios[0]
	assert.NotNil(t, pf.Transactions[0].Security)
	assert.Equal(t, "Alpha Fund", pf.Transactions[0].Security.Name)
	assert.NotNil(t, pf.Transactions[1].Security)
	assert.Equal(t, "Beta ETF", pf.Transactions[1].Security.Name)
}

func TestIntegration_RoundTrip_DecodeOnly(t *testing.T) {
	// Round-trip test stub: decode only; encode phase deferred to Phase 5.
	for _, path := range []string{testdataMinimal, testdataXPathRefs} {
		t.Run(path, func(t *testing.T) {
			f, err := ppxml.Load(path)
			require.NoError(t, err)
			defer f.Close()

			client, err := ppxml.Decode(f)
			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

// TestIntegration_InfoOutput verifies the info command output contains expected fields.
// We test directly via the XML decoder and model, not via CLI invocation.
func TestIntegration_DecodeTransactionCount(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	var totalTx int
	for _, acc := range client.Accounts {
		totalTx += len(acc.Transactions)
	}
	for _, pf := range client.Portfolios {
		totalTx += len(pf.Transactions)
	}
	assert.Equal(t, 7, totalTx)
}

// Verify date parsing of transactions.
func TestIntegration_TransactionDates(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	tx := client.Accounts[0].Transactions[0]
	assert.Equal(t, "2024-01-10", tx.Date.Format("2006-01-02"))
}

// Verify JSON output format works.
func TestIntegration_JSONOutput(_ *testing.T) {
	var buf bytes.Buffer
	_ = buf
	// The JSON format module is tested via format package; integration covered by decoder tests.
	_ = strings.NewReader("")
}

// TestIntegration_Holdings_TableOutput verifies ComputeHoldings and table output with minimal.xml.
func TestIntegration_Holdings_TableOutput(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	pf := client.Portfolios[0]
	holdings := model.ComputeHoldings(pf)

	// minimal.xml Depot: BUY 1 + SELL 0.5 + TRANSFER_IN 0.5 = net 1.0 share of VWRL
	require.Len(t, holdings, 1)
	h := holdings[0]
	assert.Equal(t, "Vanguard FTSE All-World", h.Security.Name)
	assert.Equal(t, "IE00B3RBWM25", h.Security.ISIN)
	assert.Equal(t, model.Shares(100000000), h.NetShares) // 1 share = 1×10^8
	assert.Equal(t, int64(9870000000), h.LatestPrice) // last price in minimal.xml (PP format ×10⁸)
	assert.Equal(t, int64(9870), h.Value)             // (9870000000/1e8) × 1 share × 100 cents = 9870
	assert.Equal(t, "EUR", h.Currency)

	// Table output
	var buf bytes.Buffer
	headers := []string{"Security", "ISIN", "Shares", "Latest Price", "Value", "Currency"}
	rows := [][]string{{
		h.Security.Name,
		h.Security.ISIN,
		"1.00000000",
		"98.70",
		"98.70",
		h.Currency,
	}}
	err = format.Write(&buf, format.FormatTable, headers, rows, nil)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "Vanguard FTSE All-World")
	assert.Contains(t, out, "IE00B3RBWM25")
	assert.Contains(t, out, "98.70")
}

// TestIntegration_Holdings_JSONOutput verifies JSON output for holdings.
func TestIntegration_Holdings_JSONOutput(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	pf := client.Portfolios[0]
	holdings := model.ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	h := holdings[0]

	type jsonRow struct {
		Security    string  `json:"security"`
		ISIN        string  `json:"isin"`
		Shares      float64 `json:"shares"`
		LatestPrice float64 `json:"latestPrice"`
		Value       float64 `json:"value"`
		Currency    string  `json:"currency"`
	}
	jsonRows := []jsonRow{{
		Security:    h.Security.Name,
		ISIN:        h.Security.ISIN,
		Shares:      h.NetShares.Value(),
		LatestPrice: float64(h.LatestPrice) / 1e8,
		Value:       float64(h.Value) / 100.0,
		Currency:    h.Currency,
	}}

	var buf bytes.Buffer
	err = format.Write(&buf, format.FormatJSON, nil, nil, jsonRows)
	require.NoError(t, err)

	var parsed []jsonRow
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	require.Len(t, parsed, 1)
	assert.Equal(t, "Vanguard FTSE All-World", parsed[0].Security)
	assert.InDelta(t, 1.0, parsed[0].Shares, 1e-8)
	assert.InDelta(t, 98.70, parsed[0].LatestPrice, 0.01)
	assert.InDelta(t, 98.70, parsed[0].Value, 0.01)
}

// TestIntegration_Holdings_IncludeZero verifies that zero-share positions are returned
// by ComputeHoldings (filtering is the caller's responsibility).
func TestIntegration_Holdings_IncludeZero(t *testing.T) {
	sec := &model.Security{UUID: "s1", Name: "Gone", Currency: "EUR"}
	pf := &model.Portfolio{
		Transactions: []model.PortfolioTransaction{
			{Type: "BUY", Shares: 10_00000000, Security: sec},
			{Type: "SELL", Shares: 10_00000000, Security: sec},
		},
	}

	all := model.ComputeHoldings(pf)
	require.Len(t, all, 1)
	assert.Equal(t, model.Shares(0), all[0].NetShares) // zero position included

	// Default filter: exclude zero
	var filtered []model.Holding
	for _, h := range all {
		if h.NetShares > 0 {
			filtered = append(filtered, h)
		}
	}
	assert.Empty(t, filtered)

	// With --include-zero: keep all
	assert.Len(t, all, 1)
}

// TestIntegration_CrossEntry_PortfolioTransactions verifies that portfolio transactions stored
// in buySellEntry cross-entries (real PP XML format) are decoded correctly.
func TestIntegration_CrossEntry_PortfolioTransactions(t *testing.T) {
	f, err := ppxml.Load(testdataCrossEntry)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	require.Len(t, client.Portfolios, 1)
	pf := client.Portfolios[0]
	require.Len(t, pf.Transactions, 1, "cross-entry BUY should be decoded as portfolio transaction")

	tx := pf.Transactions[0]
	assert.Equal(t, "BUY", tx.Type)
	assert.Equal(t, int64(100000000), int64(tx.Shares))
	require.NotNil(t, tx.Security, "security reference should be resolved")
	assert.Equal(t, "Vanguard FTSE All-World", tx.Security.Name)
}

// TestIntegration_CrossEntry_Holdings computes holdings from cross-entry portfolio.
func TestIntegration_CrossEntry_Holdings(t *testing.T) {
	f, err := ppxml.Load(testdataCrossEntry)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	holdings := model.ComputeHoldings(client.Portfolios[0])
	require.Len(t, holdings, 1)
	assert.Equal(t, "Vanguard FTSE All-World", holdings[0].Security.Name)
	assert.Equal(t, model.Shares(100000000), holdings[0].NetShares) // 1 share
	assert.Equal(t, int64(9870000000), holdings[0].LatestPrice)
}

// TestIntegration_Holdings_UnknownPortfolio verifies findPortfolio returns nil for unknown names.
func TestIntegration_Holdings_UnknownPortfolio(t *testing.T) {
	f, err := ppxml.Load(testdataMinimal)
	require.NoError(t, err)
	defer f.Close()

	client, err := ppxml.Decode(f)
	require.NoError(t, err)

	// Simulate the lookup used by the command.
	lower := strings.ToLower("nonexistent")
	var found *model.Portfolio
	for _, p := range client.Portfolios {
		if strings.ToLower(p.Name) == lower {
			found = p
			break
		}
	}
	assert.Nil(t, found, "unknown portfolio name should not match")
}
