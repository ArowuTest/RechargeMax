# RechargeMax Database

## Folder Structure

```
database/
│
├── tables/                  ← ONE file per table (CREATE TABLE + indexes + constraints)
│   ├── 01_admin_activity_logs.sql
│   ├── 02_admin_sessions.sql
│   ├── ... (46 tables total)
│   └── 46_wheel_prizes.sql
│
├── migrations/              ← Incremental changes ONLY (ALTER TABLE, ADD COLUMN, fixes)
│   ├── 001_rls_policies.sql
│   ├── ... (29 files)
│   └── 029_grant_all_permissions.sql
│
├── seeds/                   ← INSERT data (reference data + test data)
│   ├── 000_MASTER_PRODUCTION_SEED.sql    ← Full all-in-one production seed
│   ├── 001_comprehensive_seed_data.sql
│   ├── 002_test_data.sql                 ← Dev/staging only
│   ├── 003_test_numbers.sql
│   ├── 004_reference_data.sql
│   ├── 005_notification_templates.sql
│   ├── 006_platform_settings.sql
│   └── archived/                         ← Superseded iterations
│
├── schema.sql               ← Full pg_dump of current DB (tables + functions + triggers)
│                              Use as reference or for regenerating tables/
│
├── docker-init/             ← Shell scripts run by Docker postgres on first start
│   ├── 00_schema.sh         ← Creates all tables (reads tables/ + schema.sql)
│   └── 01_seeds.sh          ← Inserts reference data (reads seeds/004-006)
│
└── README.md
```

---

## Understanding the three layers

| Layer | Folder | Contains | Run when |
|---|---|---|---|
| **Tables** | `tables/` | `CREATE TABLE` + indexes + constraints | Once, on a **fresh** empty DB |
| **Migrations** | `migrations/` | `ALTER TABLE`, `ADD COLUMN`, constraint fixes | Against an **existing** DB to update it |
| **Seeds** | `seeds/` | `INSERT` reference data | After tables are created |

> **Rule of thumb:**  
> `tables/` = what the DB *is*  
> `migrations/` = how it *changed over time*  
> `seeds/` = what data *lives in it*

---

## Running on a fresh database

```bash
# Option 1: automated script (recommended)
./scripts/init_fresh_db.sh postgres://rechargemax:pass@localhost/rechargemax \
    --with-seeds

# Option 2: manual
psql $DB_URL -f database/schema.sql          # creates all tables + functions
./scripts/run_migrations.sh $DB_URL          # applies incremental changes
psql $DB_URL -f database/seeds/004_reference_data.sql
psql $DB_URL -f database/seeds/005_notification_templates.sql
psql $DB_URL -f database/seeds/006_platform_settings.sql
```

## Applying changes to an existing database

```bash
./scripts/run_migrations.sh $DATABASE_URL
```

## Adding a new table

1. Create `database/tables/47_your_table_name.sql` with the `CREATE TABLE` statement
2. Add a corresponding entry in `backend/migrations/052_create_your_table.sql`
3. Run the migration: `./scripts/run_migrations.sh $DATABASE_URL`
4. Regenerate `schema.sql`: `pg_dump ... --schema-only -f database/schema.sql`

## Adding a migration (changing an existing table)

1. Create `database/migrations/030_your_change.sql` with only `ALTER TABLE` statements
2. Make it **idempotent** (`IF NOT EXISTS`, `IF EXISTS`, `ON CONFLICT DO NOTHING`)
3. Mirror it in `backend/migrations/052_your_change.sql` (backend reads from there)
4. Run: `./scripts/run_migrations.sh $DATABASE_URL`

---

## Docker Compose

On `docker-compose up` (first start with empty volume):

1. `docker-init/00_schema.sh` → runs all `tables/*.sql` files, then `schema.sql`
2. `docker-init/01_seeds.sh`  → inserts reference data from `seeds/004-006`

> Re-initialise: `docker-compose down -v && docker-compose up`

---

## Notes

- `backend/migrations/` is what the Go backend runner reads (001–051 + 999)
- `database/tables/` and `database/migrations/` are the **human-readable** source of truth
- Both are kept in sync — `backend/migrations/` has the full history, `database/` has the clean split
