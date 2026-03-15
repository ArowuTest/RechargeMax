# create_views.sql — ARCHIVED

This file was a workaround created during early development when database tables
had timestamp suffixes (e.g. `admin_users_2026_01_30_14_00`).

The views mapped clean names like `admin_users` → `admin_users_2026_01_30_14_00`
so that GORM entities could find their tables without renaming every entity.

**This file is no longer needed.** The database has been fully migrated to clean
table names (`admin_users`, `transactions`, etc.) as of the schema normalisation
refactor. The GORM entities already use the correct table names via `TableName()`.

Kept here for reference only. Do not apply to any database.
