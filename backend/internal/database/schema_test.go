package database

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestRequiredMigrationsMatchFiles(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "database", "migrations"))
	if err != nil {
		t.Fatalf("read migrations directory: %v", err)
	}
	var files []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	if len(files) != len(RequiredMigrations) {
		t.Fatalf("RequiredMigrations has %d entries but database/migrations has %d .up.sql files; update schema.go", len(RequiredMigrations), len(files))
	}
	for i, name := range files {
		if RequiredMigrations[i] != name {
			t.Fatalf("RequiredMigrations[%d] = %q, want %q", i, RequiredMigrations[i], name)
		}
	}
}
