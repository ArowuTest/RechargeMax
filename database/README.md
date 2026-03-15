# RechargeMax Database

## Structure

```
database/
├── 01_admin_activity_logs.sql    ← Table definitions live HERE, directly in database/
├── 02_admin_sessions.sql
├── 03_admin_users.sql
├── ... (46 table files total)
├── 46_wheel_prizes.sql
│
├── migrations/                   ← ALTER TABLE changes only (applied to existing DBs)
│   ├── 001_rls_policies.sql
│   ├── ... (29 files)
│   └── 029_grant_all_permissions.sql
│
└── seeds/                        ← INSERT data (reference data + test data)
    ├── 000_MASTER_PRODUCTION_SEED.sql   ← Full all-in-one production seed
    ├── 001_comprehensive_seed_data.sql
    ├── 002_test_data.sql                ← Dev/staging only
    ├── 003_test_numbers.sql
    ├── 004_reference_data.sql
    ├── 005_notification_templates.sql
    └── 006_platform_settings.sql
```

---

## The three-layer rule

| What | Where | Contains |
|---|---|---|
| **Table definitions** | `database/*.sql` (root) | `CREATE TABLE`, indexes, constraints, triggers |
| **Schema changes** | `database/migrations/` | `ALTER TABLE`, `ADD COLUMN`, fixes — applied to *existing* DBs |
| **Data** | `database/seeds/` | `INSERT` statements — reference data, test data |

---

## Fresh database (first time setup)

```bash
./scripts/init_fresh_db.sh postgres://user:pass@localhost/rechargemax --with-seeds
```

Or manually:
```bash
# 1. Create tables (run all files in database/ root, in order)
for f in database/[0-9]*.sql; do psql $DB_URL -f "$f"; done

# 2. Apply incremental changes
./scripts/run_migrations.sh $DB_URL

# 3. Insert reference data
psql $DB_URL -f database/seeds/004_reference_data.sql
psql $DB_URL -f database/seeds/005_notification_templates.sql
psql $DB_URL -f database/seeds/006_platform_settings.sql
```

## Updating an existing database

```bash
# Only run new migration files — never re-run table definition files
./scripts/run_migrations.sh $DATABASE_URL
```

## Docker Compose

The `postgres` service mounts `./database` as `/docker-entrypoint-initdb.d`.  
Docker runs all `.sql` files alphabetically on first container start:
1. `01_*.sql` → `46_*.sql` — creates all tables
2. `migrations/*.sql` — skipped by Docker (subdirectories are not auto-run)
3. `seeds/*.sql` — skipped by Docker (subdirectory)

> Seeds are applied by `init_fresh_db.sh` or manually.  
> To reset: `docker-compose down -v && docker-compose up`

## Adding a new table

1. Create `database/47_your_table.sql` with the `CREATE TABLE`
2. Mirror it in `backend/migrations/052_create_your_table.sql` (same content, Go server reads from there)
3. Run on existing DBs: `psql $DB_URL -f database/47_your_table.sql`

## Adding a change to an existing table

1. Create `database/migrations/030_describe_the_change.sql` with only `ALTER TABLE`
2. Mirror it in `backend/migrations/052_describe_the_change.sql`
3. Run: `./scripts/run_migrations.sh $DB_URL`

---

## Why backend/migrations/ also exists

The Go server's manual migration runner reads from `backend/migrations/`.  
That folder contains the **full history** (CREATE TABLE + ALTER TABLE combined),  
which lets it bootstrap a database from scratch in CI and production pipelines  
without needing the `database/` folder.

`database/` = clean, human-readable, split by purpose  
`backend/migrations/` = complete ordered history for the automated runner
