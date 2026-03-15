package migrations

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed sql/*.sql sql/migrations/*.sql
var sqlFiles embed.FS

// RunAll executes all base schema SQL files then versioned migrations.
// Each STATEMENT runs independently so a failing trigger/index doesn't block table creation.
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
	totalOK, totalFail := 0, 0
	for _, f := range baseFiles {
		ok, fail := execFileByStatement(sqlDB, f)
		totalOK += ok
		totalFail += fail
	}
	log.Printf("  ✓ base schema: %d statements ok, %d warned", totalOK, totalFail)

	// Versioned migrations
	migFiles := listFiles("sql/migrations")
	log.Printf("  📋 %d migration files", len(migFiles))
	totalOK, totalFail = 0, 0
	for _, f := range migFiles {
		ok, fail := execFileByStatement(sqlDB, f)
		totalOK += ok
		totalFail += fail
	}
	log.Printf("  ✓ migrations: %d statements ok, %d warned", totalOK, totalFail)

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

// splitStatements splits a SQL file into individual statements.
// It handles $$ dollar-quoted blocks (used in PL/pgSQL functions).
func splitStatements(sql string) []string {
	// Replace line comments
	reLineComment := regexp.MustCompile(`--[^\n]*`)
	cleaned := reLineComment.ReplaceAllString(sql, "")

	var stmts []string
	var buf strings.Builder
	inDollarQuote := false
	dollarTag := ""
	i := 0
	runes := []rune(cleaned)

	for i < len(runes) {
		// Check for dollar-quote start/end
		if !inDollarQuote {
			// Look for $tag$ pattern
			if runes[i] == '$' {
				end := strings.Index(string(runes[i+1:]), "$")
				if end >= 0 {
					tag := "$" + string(runes[i+1:i+1+end]) + "$"
					inDollarQuote = true
					dollarTag = tag
					buf.WriteString(tag)
					i += len([]rune(tag))
					continue
				}
			}
			if runes[i] == ';' {
				stmt := strings.TrimSpace(buf.String())
				if stmt != "" {
					stmts = append(stmts, stmt)
				}
				buf.Reset()
				i++
				continue
			}
		} else {
			// Check if we're closing the dollar quote
			remaining := string(runes[i:])
			if strings.HasPrefix(remaining, dollarTag) {
				buf.WriteString(dollarTag)
				i += len([]rune(dollarTag))
				inDollarQuote = false
				dollarTag = ""
				continue
			}
		}
		buf.WriteRune(runes[i])
		i++
	}

	// Last statement without semicolon
	if stmt := strings.TrimSpace(buf.String()); stmt != "" {
		stmts = append(stmts, stmt)
	}

	return stmts
}

// execFileByStatement executes each statement in the SQL file independently.
// Returns (successCount, failureCount).
func execFileByStatement(db *sql.DB, path string) (int, int) {
	data, err := sqlFiles.ReadFile(path)
	if err != nil {
		log.Printf("  ❌ Cannot read %s: %v", filepath.Base(path), err)
		return 0, 1
	}

	stmts := splitStatements(string(data))
	ok, fail := 0, 0
	for _, stmt := range stmts {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			// Only log non-trivial errors (ignore "already exists" etc.)
			errStr := err.Error()
			if !strings.Contains(errStr, "already exists") &&
				!strings.Contains(errStr, "duplicate key") &&
				!strings.Contains(errStr, "does not exist") {
				log.Printf("  ⚠️  %s: %v", filepath.Base(path), err)
			}
			fail++
		} else {
			ok++
		}
	}
	return ok, fail
}
