package model

import "time"

// Account represents a cash/deposit account.
type Account struct {
	UUID         string
	Name         string
	Currency     string
	Note         string
	Retired      bool
	Transactions []AccountTransaction
}

// AccountTransaction represents a single transaction within an account.
type AccountTransaction struct {
	UUID        string
	Type        string
	Date        time.Time
	Amount      int64  // integer minor units
	Currency    string
	Note        string
	SecurityRef string // raw reference attribute (resolved to Security pointer)
	Security    *Security
	Units       []TxUnit
}

// TxUnit represents a sub-unit of a transaction (e.g., fees, taxes).
type TxUnit struct {
	Type     string
	Amount   int64
	Currency string
	Shares   Shares
}
