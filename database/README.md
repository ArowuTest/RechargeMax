# RechargeMax Database

## Folder Structure

```
database/
├── schema.sql              ← FULL current schema (all CREATE TABLE, functions, triggers)
│                             Generated from: pg_dump --schema-only
│                             Run this on a FRESH database before anything else.
│
├── migrations/             ← INCREMENTAL changes only (ALTER TABLE, ADD COLUMN, fixes)
│   ├── 001_rls_policies.sql
│   ├── 002_points_adjustments.sql
│   ├── ...
│   └── 029_grant_all_permissions.sql
│
├── seeds/                  ← Reference & test data (INSERT statements)
│   ├── 001_comprehensive_seed_data.sql   ← Legacy comprehensive seed
│   ├── 002_test_data.sql                 ← Test users + transactions (dev/staging only)
│   ├── 003_test_numbers.sql              ← Nigerian test MSISDN numbers
│   ├── 004_reference_data.sql            ← Networks, data plans, subscription tiers
│   ├── 005_notification_templates.sql    ← Notification template definitions
│   ├── 006_platform_settings.sql         ← Platform configuration key/value pairs
│   ├── MASTER_PRODUCTION_SEED_CORRECTED.sql  ← Full production seed (all-in-one)
│   └── archived/                         ← Superseded seed iterations (do not run)
│
├── docker-init/            ← Shell scripts mounted into Docker postgres initdb.d
│   ├── 00_schema.sh        ← Runs schema.sql on first container start
│   └── 01_seeds.sh         ← Runs reference seeds on first container start
│
└── README.md               ← This file
```

---

## When to use each file

| Situation | What to run |
|---|---|
| **Fresh database** (dev, CI, staging) | `scripts/init_fresh_db.sh` |
| **Existing database** (apply changes) | `scripts/run_migrations.sh` |
| **Docker Compose first start** | Automatic via `docker-init/` |
| **Production seed data** | `seeds/MASTER_PRODUCTION_SEED_CORRECTED.sql` |
| **Test/dev seed data** | `seeds/002_test_data.sql` + `seeds/003_test_numbers.sql` |

---

## Quickstart (local dev)

```bash
# Fresh local database
createdb rechargemax
./scripts/init_fresh_db.sh postgres://rechargemax:rechargemax@localhost/rechargemax \
    --with-seeds --with-test-data
```

---

## Docker Compose

The `postgres` service in `docker-compose.yml` mounts:

| Mount | Purpose |
|---|---|
| `database/docker-init/` → `/docker-entrypoint-initdb.d/` | Runs `00_schema.sh` then `01_seeds.sh` on **first init only** |
| `database/schema.sql` → `/docker-entrypoint-initdb.d/schema.sql` | Schema file referenced by `00_schema.sh` |
| `database/seeds/` → `/docker-entrypoint-initdb.d/seeds/` | Seed files referenced by `01_seeds.sh` |

> ⚠️ `docker-entrypoint-initdb.d` scripts only run when the data volume is **empty** (first start).
> To re-initialise: `docker-compose down -v && docker-compose up`

---

## Adding a new migration

1. Create `database/migrations/030_your_description.sql`
2. Write only `ALTER TABLE` / `ADD COLUMN` / `DROP CONSTRAINT` statements
3. Make it **idempotent** (`IF NOT EXISTS`, `IF EXISTS`, `ON CONFLICT DO NOTHING`)
4. Run: `./scripts/run_migrations.sh $DATABASE_URL`
5. After running on all environments, regenerate schema: `pg_dump ... -f database/schema.sql`

---

## Migration naming convention

```
NNN_descriptive_name.sql
```

| Range | Purpose |
|---|---|
| `001` – `029` | Current incremental migrations |
| `030` + | Future migrations (add here) |

**Never rename or renumber** migrations that have already been applied to any environment.
