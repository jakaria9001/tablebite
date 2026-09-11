package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationVersion(t *testing.T) {
	for name, want := range map[string]int{
		"001_core.sql":    1,
		"012_example.sql": 12,
	} {
		got, err := migrationVersion(name)
		if err != nil || got != want {
			t.Fatalf("migrationVersion(%q) = %d, %v; want %d", name, got, err, want)
		}
	}
	for _, name := range []string{"core.sql", "abc_core.sql", "001"} {
		if _, err := migrationVersion(name); err == nil {
			t.Fatalf("migrationVersion(%q) should fail", name)
		}
	}
}

func TestMigrationFilesAreNumberedAndContainSQL(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected migration files")
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		if _, err := migrationVersion(entry.Name()); err != nil {
			t.Fatalf("invalid migration filename %q: %v", entry.Name(), err)
		}
		data, err := os.ReadFile(filepath.Join("..", "..", "migrations", entry.Name()))
		if err != nil {
			t.Fatalf("read migration %q: %v", entry.Name(), err)
		}
		if strings.TrimSpace(string(data)) == "" {
			t.Fatalf("migration %q is empty", entry.Name())
		}
	}
}
