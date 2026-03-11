package xml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const xpathRefFixture = `<?xml version="1.0" encoding="UTF-8"?>
<client>
  <version>52</version>
  <baseCurrency>EUR</baseCurrency>
  <securities>
    <security>
      <uuid>sec-a</uuid>
      <name>Security A</name>
      <currencyCode>EUR</currencyCode>
      <prices/>
      <events/>
    </security>
    <security>
      <uuid>sec-b</uuid>
      <name>Security B</name>
      <currencyCode>USD</currencyCode>
      <prices/>
      <events/>
    </security>
    <security>
      <uuid>sec-c</uuid>
      <name>Security C</name>
      <currencyCode>GBP</currencyCode>
      <prices/>
      <events/>
    </security>
  </securities>
  <accounts>
    <account>
      <uuid>acc-1</uuid>
      <name>Account 1</name>
      <currencyCode>EUR</currencyCode>
      <transactions>
        <account-transaction>
          <uuid>t1</uuid>
          <date>2024-01-01T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>1000</amount>
          <type>BUY</type>
          <security reference="../../../../../securities/security[3]"/>
        </account-transaction>
        <account-transaction>
          <uuid>t2</uuid>
          <date>2024-02-01T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>2000</amount>
          <type>BUY</type>
          <security reference="../../../../../securities/security[1]"/>
        </account-transaction>
      </transactions>
    </account>
  </accounts>
  <portfolios>
    <portfolio>
      <uuid>pf-1</uuid>
      <name>Portfolio 1</name>
      <transactions>
        <portfolio-transaction>
          <uuid>pt1</uuid>
          <date>2024-01-01T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>1000</amount>
          <shares>100000000</shares>
          <type>BUY</type>
          <security reference="../../../../../securities/security[2]"/>
        </portfolio-transaction>
      </transactions>
    </portfolio>
  </portfolios>
</client>`

func TestResolveXPathReferences(t *testing.T) {
	client, err := Decode(strings.NewReader(xpathRefFixture))
	require.NoError(t, err)

	// First account transaction references security[3] = "Security C"
	tx1 := client.Accounts[0].Transactions[0]
	require.NotNil(t, tx1.Security, "security reference should be resolved")
	assert.Equal(t, "Security C", tx1.Security.Name)

	// Second account transaction references security[1] = "Security A"
	tx2 := client.Accounts[0].Transactions[1]
	require.NotNil(t, tx2.Security)
	assert.Equal(t, "Security A", tx2.Security.Name)

	// Portfolio transaction references security[2] = "Security B"
	ptx := client.Portfolios[0].Transactions[0]
	require.NotNil(t, ptx.Security)
	assert.Equal(t, "Security B", ptx.Security.Name)
}

const idModeFixture = `<?xml version="1.0" encoding="UTF-8"?>
<client id="root">
  <version>52</version>
  <baseCurrency>EUR</baseCurrency>
  <securities>
    <security id="sec-id-1">
      <uuid>sec-id-1</uuid>
      <name>ID Mode Security</name>
      <currencyCode>EUR</currencyCode>
      <prices/>
      <events/>
    </security>
  </securities>
  <accounts>
    <account id="acc-id-1">
      <uuid>acc-id-1</uuid>
      <name>ID Account</name>
      <currencyCode>EUR</currencyCode>
      <transactions>
        <account-transaction>
          <uuid>tx-id-1</uuid>
          <date>2024-01-01T00:00</date>
          <currencyCode>EUR</currencyCode>
          <amount>500</amount>
          <type>BUY</type>
          <security reference="sec-id-1"/>
        </account-transaction>
      </transactions>
    </account>
  </accounts>
  <portfolios/>
</client>`

func TestResolveIDModeReferences(t *testing.T) {
	client, err := Decode(strings.NewReader(idModeFixture))
	require.NoError(t, err)

	tx := client.Accounts[0].Transactions[0]
	require.NotNil(t, tx.Security, "ID-mode reference should be resolved")
	assert.Equal(t, "ID Mode Security", tx.Security.Name)
}
