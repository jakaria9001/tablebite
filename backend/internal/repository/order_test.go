package repository

import "testing"

func TestPriceForVariantUsesAuthoritativeVariantPrice(t *testing.T) {
	regular, half, full := 100.0, 60.0, 110.0
	tests := []struct {
		variant string
		want    *float64
	}{
		{"regular", &regular},
		{"half", &half},
		{"full", &full},
		{"invalid", nil},
	}
	for _, tt := range tests {
		got := priceForVariant(&regular, &half, &full, tt.variant)
		if (got == nil) != (tt.want == nil) {
			t.Fatalf("variant %q returned %v, want %v", tt.variant, got, tt.want)
		}
		if got != nil && *got != *tt.want {
			t.Fatalf("variant %q price = %v, want %v", tt.variant, *got, *tt.want)
		}
	}
}

func TestMoneyRoundsToCurrencyPrecision(t *testing.T) {
	if got := money(10.125); got != 10.13 {
		t.Fatalf("money(10.125) = %v, want 10.13", got)
	}
}
