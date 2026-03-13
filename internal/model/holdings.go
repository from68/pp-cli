package model

import "math"

// ComputeHoldings aggregates net share positions from a portfolio's transactions.
// It returns one Holding per security that appears in the portfolio.
// Securities with net shares ≤ 0 are included; callers filter as needed.
func ComputeHoldings(portfolio *Portfolio) []Holding {
	type entry struct {
		sec    *Security
		shares Shares
	}

	// Preserve first-seen order.
	var order []string
	byUUID := map[string]*entry{}

	for _, tx := range portfolio.Transactions {
		if tx.Security == nil {
			continue
		}
		sec := tx.Security
		e, ok := byUUID[sec.UUID]
		if !ok {
			e = &entry{sec: sec}
			byUUID[sec.UUID] = e
			order = append(order, sec.UUID)
		}
		switch tx.Type {
		case "BUY", "DELIVERY_IN", "TRANSFER_IN":
			e.shares += tx.Shares
		case "SELL", "DELIVERY_OUT", "TRANSFER_OUT":
			e.shares -= tx.Shares
		}
	}

	holdings := make([]Holding, 0, len(order))
	for _, uuid := range order {
		e := byUUID[uuid]
		h := Holding{
			Security:  e.sec,
			NetShares: e.shares,
			Currency:  e.sec.Currency,
		}
		if len(e.sec.Prices) > 0 {
			h.LatestPrice = e.sec.Prices[len(e.sec.Prices)-1].Value
			// Price stored as price × 10⁸; value in minor units (cents).
			h.Value = int64(math.Round(float64(h.LatestPrice) / 1e8 * e.shares.Value() * 100))
		}
		holdings = append(holdings, h)
	}
	return holdings
}
