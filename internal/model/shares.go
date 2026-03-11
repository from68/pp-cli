package model

// Shares represents a share quantity stored as int64 multiples of 10^8.
// e.g., 50000000000 = 500 shares.
type Shares int64

// Value returns the shares as a float64 decimal value.
func (s Shares) Value() float64 {
	return float64(s) / 1e8
}
