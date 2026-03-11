package model

import "time"

// Security represents a financial security (stock, ETF, bond, etc.).
type Security struct {
	UUID      string
	Name      string
	Currency  string
	ISIN      string
	Ticker    string
	Feed      string
	Retired   bool
	UpdatedAt time.Time
	Prices    []SecurityPrice
	Events    []SecurityEvent
}

// SecurityPrice represents a price data point for a security.
type SecurityPrice struct {
	Date  time.Time
	Value int64 // integer minor units
}

// SecurityEvent represents a corporate action or dividend event.
type SecurityEvent struct {
	Type    string
	Date    time.Time
	Amount  int64
	Details string
}
