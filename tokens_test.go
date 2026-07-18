package main

import "testing"

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"a", 1},        // rounds up
		{"abcd", 1},     // exactly 4 chars
		{"abcde", 2},    // 5 chars -> ceil(5/4)
		{"abcdefgh", 2}, // 8 chars
	}
	for _, tt := range tests {
		if got := estimateTokens(tt.in); got != tt.want {
			t.Errorf("estimateTokens(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestHumanCount(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{5, "5"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{1000000, "1,000,000"},
		{-2500, "-2,500"},
	}
	for _, tt := range tests {
		if got := humanCount(tt.in); got != tt.want {
			t.Errorf("humanCount(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
