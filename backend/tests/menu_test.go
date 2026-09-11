package repository_test

import (
	"testing"

	"github.com/jakaria9001/tablebite/internal/repository"
)

func TestMenuItemCategoryColumnName(t *testing.T) {
	tests := []struct {
		name                string
		hasCategoryID       bool
		hasCategorySourceID bool
		want                string
		wantErr             bool
	}{
		{
			name:                "prefer legacy category_id when present",
			hasCategoryID:       true,
			hasCategorySourceID: true,
			want:                "category_id",
		},
		{
			name:                "fallback to category_source_id",
			hasCategoryID:       false,
			hasCategorySourceID: true,
			want:                "category_source_id",
		},
		{
			name:    "error when neither exists",
			wantErr: true,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repository.MenuItemCategoryColumnName(tt.hasCategoryID, tt.hasCategorySourceID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
