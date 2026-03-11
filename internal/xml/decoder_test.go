package xml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const simpleFixture = `<?xml version="1.0" encoding="UTF-8"?>
<client>
  <version>52</version>
  <baseCurrency>EUR</baseCurrency>
  <securities>
    <security>
      <uuid>sec-uuid-1</uuid>
      <name>Vanguard FTSE All-World</name>
      <currencyCode>EUR</currencyCode>
      <isin>IE00B3RBWM25</isin>
      <tickerSymbol>VWRL</tickerSymbol>
      <prices>
        <price t="2024-01-02" v="12345"/>
        <price t="2024-01-03" v="12400"/>
      </prices>
      <events>
        <event>
          <type>STOCK_SPLIT</type>
          <date>2023-06-01</date>
          <amount>0</amount>
        </event>
        <dividendEvent>
          <type>DIVIDEND</type>
          <date>2023-09-15</date>
          <amount>500</amount>
        </dividendEvent>
      </events>
    </security>
    <security>
      <uuid>sec-uuid-2</uuid>
      <name>iShares Core MSCI World</name>
      <currencyCode>USD</currencyCode>
      <isin>IE00B4L5Y983</isin>
      <isRetired>true</isRetired>
      <prices/>
    </security>
  </securities>
  <accounts>
    <account>
      <uuid>acc-uuid-1</uuid>
      <name>Checking Account</name>
      <currencyCode>EUR</currencyCode>
      <transactions>
        <account-transaction>
          <uuid>tx-uuid-1</uuid>
          <date>2024-01-15T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>100000</amount>
          <type>DEPOSIT</type>
          <note>Initial deposit</note>
        </account-transaction>
        <account-transaction>
          <uuid>tx-uuid-2</uuid>
          <date>2024-01-20T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>12345</amount>
          <type>BUY</type>
          <security reference="../../../../../securities/security[1]"/>
        </account-transaction>
      </transactions>
    </account>
  </accounts>
  <portfolios>
    <portfolio>
      <uuid>pf-uuid-1</uuid>
      <name>My Depot</name>
      <transactions>
        <portfolio-transaction>
          <uuid>ptx-uuid-1</uuid>
          <date>2024-01-20T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>12345</amount>
          <shares>100000000</shares>
          <type>BUY</type>
          <security reference="../../../../../securities/security[1]"/>
        </portfolio-transaction>
      </transactions>
    </portfolio>
  </portfolios>
</client>`

func TestDecode_BasicClient(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	assert.Equal(t, 52, client.Version)
	assert.Equal(t, "EUR", client.BaseCurrency)
}

func TestDecode_Securities(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	require.Len(t, client.Securities, 2)
	s := client.Securities[0]
	assert.Equal(t, "sec-uuid-1", s.UUID)
	assert.Equal(t, "Vanguard FTSE All-World", s.Name)
	assert.Equal(t, "EUR", s.Currency)
	assert.Equal(t, "IE00B3RBWM25", s.ISIN)
	assert.Equal(t, "VWRL", s.Ticker)
	assert.False(t, s.Retired)

	require.Len(t, s.Prices, 2)
	assert.Equal(t, int64(12345), s.Prices[0].Value)
	assert.Equal(t, int64(12400), s.Prices[1].Value)

	// Retired security
	s2 := client.Securities[1]
	assert.True(t, s2.Retired)
	assert.Empty(t, s2.Prices)
}

func TestDecode_Events(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	s := client.Securities[0]
	require.Len(t, s.Events, 2)
	assert.Equal(t, "event", s.Events[0].Type)
	assert.Equal(t, "dividendEvent", s.Events[1].Type)
	assert.Equal(t, int64(500), s.Events[1].Amount)
}

func TestDecode_AccountTransactions(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	require.Len(t, client.Accounts, 1)
	acc := client.Accounts[0]
	assert.Equal(t, "Checking Account", acc.Name)

	require.Len(t, acc.Transactions, 2)
	tx := acc.Transactions[0]
	assert.Equal(t, "DEPOSIT", tx.Type)
	assert.Equal(t, int64(100000), tx.Amount)
	assert.Equal(t, "Initial deposit", tx.Note)
}

func TestDecode_AmountsStoredAsInteger(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	tx := client.Accounts[0].Transactions[0]
	assert.Equal(t, int64(100000), tx.Amount)
}

func TestDecode_SharesStoredAsInteger(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	ptx := client.Portfolios[0].Transactions[0]
	assert.Equal(t, int64(100000000), int64(ptx.Shares))
}

func TestDecode_PortfolioTransactions(t *testing.T) {
	client, err := Decode(strings.NewReader(simpleFixture))
	require.NoError(t, err)

	require.Len(t, client.Portfolios, 1)
	pf := client.Portfolios[0]
	require.Len(t, pf.Transactions, 1)
	ptx := pf.Transactions[0]
	assert.Equal(t, "BUY", ptx.Type)
}
