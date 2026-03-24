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
//
// BASE SCHEMA (sql/*.sql): always re-run on every deploy — all statements are
// CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT EXISTS / etc., so they are
// fully idempotent.
//
// VERSIONED MIGRATIONS (sql/migrations/*.sql): tracked in the schema_migrations
// table.  Each file is run EXACTLY ONCE; subsequent deploys skip already-applied
// files.  This prevents destructive statements like TRUNCATE from running again.
func RunAll(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("❌ migrations.RunAll: cannot get sql.DB: %v", err)
		return
	}

	log.Println("📦 Running embedded SQL migrations...")

	// ── 1. Ensure the migration-tracking table exists ──────────────────────
	_, _ = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     VARCHAR(255) PRIMARY KEY,
		applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)

	// ── 2. Base schema (always idempotent — run every deploy) ──────────────
	baseFiles := listFiles("sql")
	log.Printf("  📋 %d base schema files", len(baseFiles))
	totalOK, totalFail := 0, 0
	for _, f := range baseFiles {
		ok, fail := execFileByStatement(sqlDB, f)
		totalOK += ok
		totalFail += fail
	}
	log.Printf("  ✓ base schema: %d statements ok, %d warned", totalOK, totalFail)

	// ── 3. Versioned migrations (run each file exactly once) ────────────────
	migFiles := listFiles("sql/migrations")
	log.Printf("  📋 %d versioned migration files found", len(migFiles))

	applied, skipped := 0, 0
	for _, f := range migFiles {
		version := filepath.Base(f) // e.g. "049_clean_and_reseed_wheel_prizes.sql"

		// Check if already applied
		var count int
		row := sqlDB.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, version)
		if err := row.Scan(&count); err != nil || count > 0 {
			skipped++
			continue
		}

		// Run the migration
		ok, fail := execFileByStatement(sqlDB, f)
		if ok > 0 || fail == 0 {
			// Record as applied even if some statements warned — the file ran
			if _, err := sqlDB.Exec(`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, version); err != nil {
				log.Printf("  ⚠️  Could not record migration %s: %v", version, err)
			} else {
				log.Printf("  ✅ Applied migration: %s (%d ok, %d warned)", version, ok, fail)
				applied++
			}
		} else {
			log.Printf("  ❌ Migration %s failed (%d ok, %d failed)", version, ok, fail)
		}
	}
	log.Printf("  ✓ versioned migrations: %d newly applied, %d already done", applied, skipped)

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
			// Suppress expected idempotency errors — these occur on re-runs when
			// tables/indexes/constraints already exist from a previous deploy.
			errStr := err.Error()
			if !strings.Contains(errStr, "already exists") &&
				!strings.Contains(errStr, "duplicate key") &&
				!strings.Contains(errStr, "does not exist") &&
				!strings.Contains(errStr, "multiple primary keys") {
				log.Printf("  ⚠️  %s: %v", filepath.Base(path), err)
			}
			fail++
		} else {
			ok++
		}
	}
	return ok, fail
}
