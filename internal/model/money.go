package model

// Money represents a monetary value stored as integer minor units (e.g., 100 = €1.00).
type Money struct {
	Amount   int64
	Currency string
}

// Value returns the money amount as a float64 decimal value.
func (m Money) Value() float64 {
	return float64(m.Amount) / 100.0
}
