package repository

import "testing"

func TestNormalizeReviewRating(t *testing.T) {
	tests := []struct {
		name   string
		rating int
		want   int
	}{
		{name: "below range clamps to minimum", rating: 0, want: 1},
		{name: "above range clamps to maximum", rating: 6, want: 5},
		{name: "within range stays unchanged", rating: 4, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeReviewRating(tt.rating); got != tt.want {
				t.Fatalf("normalizeReviewRating(%d) = %d, want %d", tt.rating, got, tt.want)
			}
		})
	}
}
