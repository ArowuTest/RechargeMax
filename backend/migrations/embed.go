package migrations

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed sql/*.sql sql/migrations/*.sql
var sqlFiles embed.FS

// RunAll executes all base schema SQL files then versioned migrations.
// Errors are logged but do not abort: each file is best-effort (idempotent SQL).
func RunAll(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("❌ migrations.RunAll: cannot get sql.DB: %v", err)
		return
	}

	log.Println("📦 Running embedded SQL migrations...")

	// Base schema files
	for _, f := range listFiles("sql") {
		execFile(sqlDB, f)
	}

	// Versioned migrations
	for _, f := range listFiles("sql/migrations") {
		execFile(sqlDB, f)
	}

	log.Println("📦 Migrations complete")
}

func listFiles(dir string) []string {
	entries, err := fs.ReadDir(sqlFiles, dir)
	if err != nil {
		log.Printf("  ⚠️  Cannot read dir %s: %v", dir, err)
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	log.Printf("  ℹ️  Found %d files in %s", len(files), dir)
	return files
}

func execFile(db *sql.DB, path string) {
	data, err := sqlFiles.ReadFile(path)
	if err != nil {
		log.Printf("  ❌ Cannot read %s: %v", filepath.Base(path), err)
		return
	}
	if _, err := db.Exec(string(data)); err != nil {
		log.Printf("  ⚠️  %s: %v", filepath.Base(path), err)
	} else {
		log.Printf("  ✓ %s", filepath.Base(path))
	}
}
