package model

// Holding represents a net share position in a portfolio for a single security.
type Holding struct {
	Security    *Security
	NetShares   Shares
	LatestPrice int64  // 0 if unknown
	Currency    string
	Value       int64  // 0 if unknown; minor units
}
