package repository_test

import (
	"strings"
	"testing"

	"github.com/jakaria9001/tablebite/internal/repository"
)

func TestFormatTableNumber(t *testing.T) {
	got := repository.FormatTableNumber(1)
	if got != "01" {
		t.Fatalf("expected 01, got %q", got)
	}

	got = repository.FormatTableNumber(12)
	if got != "12" {
		t.Fatalf("expected 12, got %q", got)
	}
}

func TestGenerateTableToken(t *testing.T) {
	token := repository.GenerateTableToken(7, 42)
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if strings.Contains(token, " ") {
		t.Fatalf("expected token without spaces, got %q", token)
	}
	if !strings.HasPrefix(token, "tbl-") {
		t.Fatalf("expected token prefix tbl-, got %q", token)
	}
}
