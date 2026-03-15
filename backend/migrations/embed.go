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
// Each file runs in its own transaction so failures don't block subsequent files.
func RunAll(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("❌ migrations.RunAll: cannot get sql.DB: %v", err)
		return
	}

	log.Println("📦 Running embedded SQL migrations...")

	// Base schema files (creates tables, indexes, triggers)
	baseFiles := listFiles("sql")
	log.Printf("  📋 %d base schema files", len(baseFiles))
	ok, fail := 0, 0
	for _, f := range baseFiles {
		if err := execFileInTx(sqlDB, f); err != nil {
			log.Printf("  ⚠️  %s: %v", filepath.Base(f), err)
			fail++
		} else {
			ok++
		}
	}
	log.Printf("  ✓ base schema: %d ok, %d warned", ok, fail)

	// Versioned migrations (alters, seeds, indexes)
	migFiles := listFiles("sql/migrations")
	log.Printf("  📋 %d migration files", len(migFiles))
	ok, fail = 0, 0
	for _, f := range migFiles {
		if err := execFileInTx(sqlDB, f); err != nil {
			log.Printf("  ⚠️  %s: %v", filepath.Base(f), err)
			fail++
		} else {
			ok++
		}
	}
	log.Printf("  ✓ migrations: %d ok, %d warned", ok, fail)

	log.Println("📦 Migrations complete")
}

func listFiles(dir string) []string {
	entries, err := fs.ReadDir(sqlFiles, dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	return files
}

// execFileInTx runs a SQL file inside a new transaction.
// If the file fails, the transaction is rolled back (so DB stays clean).
func execFileInTx(db *sql.DB, path string) error {
	data, err := sqlFiles.ReadFile(path)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(data)); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
