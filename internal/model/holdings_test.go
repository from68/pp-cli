package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSecurity(uuid, name, currency string, prices ...int64) *Security {
	sec := &Security{UUID: uuid, Name: name, Currency: currency}
	for i, v := range prices {
		sec.Prices = append(sec.Prices, SecurityPrice{
			Date:  time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			Value: v,
		})
	}
	return sec
}

func makeTx(txType string, shares int64, sec *Security) PortfolioTransaction {
	return PortfolioTransaction{
		Type:     txType,
		Shares:   Shares(shares),
		Security: sec,
	}
}

func TestComputeHoldings_BuyOnly(t *testing.T) {
	sec := makeSecurity("s1", "Alpha", "EUR", 100_00000000) // $100.00 in PP format (×10⁸)
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("BUY", 100_00000000, sec), // 100 shares
		},
	}

	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	assert.Equal(t, Shares(100_00000000), holdings[0].NetShares)
	assert.Equal(t, int64(100_00000000), holdings[0].LatestPrice)
	// Value = (100_00000000/1e8) × 100 shares × 100 cents = 100.00 × 100 × 100 = 1000000 cents
	assert.Equal(t, int64(1000000), holdings[0].Value)
}

func TestComputeHoldings_BuyAndSell(t *testing.T) {
	sec := makeSecurity("s1", "Alpha", "EUR", 100_00000000)
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("BUY", 100_00000000, sec),  // +100
			makeTx("SELL", 40_00000000, sec),  // -40
		},
	}

	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	assert.Equal(t, Shares(60_00000000), holdings[0].NetShares)
}

func TestComputeHoldings_DeliveryIn(t *testing.T) {
	sec := makeSecurity("s1", "Alpha", "EUR")
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("DELIVERY_IN", 50_00000000, sec),
		},
	}

	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	assert.Equal(t, Shares(50_00000000), holdings[0].NetShares)
	assert.Equal(t, int64(0), holdings[0].LatestPrice)
	assert.Equal(t, int64(0), holdings[0].Value)
}

func TestComputeHoldings_DeliveryOut(t *testing.T) {
	sec := makeSecurity("s1", "Alpha", "EUR")
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("DELIVERY_IN", 50_00000000, sec),
			makeTx("DELIVERY_OUT", 50_00000000, sec),
		},
	}

	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	assert.Equal(t, Shares(0), holdings[0].NetShares)
}

func TestComputeHoldings_ZeroPositionIncluded(t *testing.T) {
	sec := makeSecurity("s1", "Alpha", "EUR")
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("BUY", 10_00000000, sec),
			makeTx("SELL", 10_00000000, sec),
		},
	}

	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	assert.Equal(t, Shares(0), holdings[0].NetShares) // included; caller filters
}

func TestComputeHoldings_MixedTypes(t *testing.T) {
	sec := makeSecurity("s1", "Alpha", "EUR", 50_00000000)
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("BUY", 100_00000000, sec),       // +100
			makeTx("SELL", 20_00000000, sec),       // -20
			makeTx("DELIVERY_IN", 10_00000000, sec), // +10
			makeTx("TRANSFER_OUT", 5_00000000, sec), // -5
			makeTx("TRANSFER_IN", 15_00000000, sec), // +15
		},
	}

	// Net = 100 - 20 + 10 - 5 + 15 = 100
	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 1)
	assert.Equal(t, Shares(100_00000000), holdings[0].NetShares)
}

func TestComputeHoldings_MultipleSecurities(t *testing.T) {
	secA := makeSecurity("s1", "Alpha", "EUR", 100_00000000)
	secB := makeSecurity("s2", "Beta", "USD", 200_00000000)
	pf := &Portfolio{
		Transactions: []PortfolioTransaction{
			makeTx("BUY", 10_00000000, secA),
			makeTx("BUY", 5_00000000, secB),
		},
	}

	holdings := ComputeHoldings(pf)
	require.Len(t, holdings, 2)
	assert.Equal(t, "Alpha", holdings[0].Security.Name)
	assert.Equal(t, "Beta", holdings[1].Security.Name)
}
