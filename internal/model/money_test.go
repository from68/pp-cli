package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoneyValue(t *testing.T) {
	tests := []struct {
		amount   int64
		expected float64
	}{
		{0, 0.0},
		{100, 1.0},
		{12345, 123.45},
		{1, 0.01},
		{-500, -5.0},
	}
	for _, tt := range tests {
		m := Money{Amount: tt.amount}
		assert.InDelta(t, tt.expected, m.Value(), 1e-9)
	}
}

func TestSharesValue(t *testing.T) {
	tests := []struct {
		shares   Shares
		expected float64
	}{
		{0, 0.0},
		{100000000, 1.0},
		{50000000000, 500.0},
		{150000000, 1.5},
	}
	for _, tt := range tests {
		assert.InDelta(t, tt.expected, tt.shares.Value(), 1e-9)
	}
}
