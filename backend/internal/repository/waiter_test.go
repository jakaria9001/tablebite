package repository

import "testing"

func TestValidStaffTransition(t *testing.T) {
	tests := []struct {
		current string
		next    string
		want    bool
	}{
		{"new", "accepted", true},
		{"new", "cancelled", true},
		{"new", "preparing", false},
		{"accepted", "preparing", true},
		{"preparing", "ready", true},
		{"ready", "served", true},
		{"served", "completed", true},
		{"completed", "new", false},
		{"cancelled", "accepted", false},
	}
	for _, tt := range tests {
		if got := validStaffTransition(tt.current, tt.next); got != tt.want {
			t.Fatalf("transition %q -> %q = %v, want %v", tt.current, tt.next, got, tt.want)
		}
	}
}
