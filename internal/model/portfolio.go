package model

import "time"

// Portfolio represents a securities portfolio (depot).
type Portfolio struct {
	UUID             string
	Name             string
	ReferenceAccount *Account
	Transactions     []PortfolioTransaction
}

// PortfolioTransaction represents a buy/sell or transfer within a portfolio.
type PortfolioTransaction struct {
	UUID        string
	Type        string
	Date        time.Time
	Shares      Shares // integer ×10^8
	Amount      int64  // integer minor units
	Currency    string
	Note        string
	SecurityRef string // raw reference attribute (resolved to Security pointer)
	Security    *Security
	Units       []TxUnit
}
