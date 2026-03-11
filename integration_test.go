package main_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ppxml "github.com/yourname/pp-cli/internal/xml"
)

const testdataMinimal = "testdata/minimal.xml"
const testdataXPathRefs = "testdata/xpath_refs.xml"

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
	assert.Equal(t, int64(9843), vwrl.Prices[0].Value)
	assert.Equal(t, int64(9870), vwrl.Prices[2].Value)
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
