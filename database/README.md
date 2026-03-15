# RechargeMax Database

## Structure

```
database/
├── migrations/          ← All 52 SQL migrations (canonical copy, mirrors backend/migrations/)
│   ├── 001_core_tables_schema.sql
│   ├── ...
│   └── 049_transaction_limits.sql
├── seeds/
│   ├── 001_comprehensive_seed_data.sql   ← Core reference data (networks, plans, tiers, prizes)
│   ├── 002_test_data.sql                 ← Test users, transactions (dev/staging only)
│   ├── MASTER_PRODUCTION_SEED_CORRECTED.sql  ← Full production seed (schema-aligned)
│   ├── test_numbers_seed.sql             ← Nigerian test MSISDN numbers
│   └── archived/                         ← Superseded seed iterations (do not run)
└── README.md
```

## Running Migrations

```bash
# Using the backend runner (recommended)
cd backend
for f in migrations/*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

Or use the automated migration runner built into the Go server startup.

## Running Seeds

**Development / Staging:**
```bash
psql "$DATABASE_URL" -f database/seeds/001_comprehensive_seed_data.sql
psql "$DATABASE_URL" -f database/seeds/002_test_data.sql
```

**Production:**
```bash
psql "$DATABASE_URL" -f database/seeds/MASTER_PRODUCTION_SEED_CORRECTED.sql
```

## Migration Naming Convention

| Pattern | Description |
|---|---|
| `001_` – `049_` | Sequential numbered migrations |
| `20260223HHMMSS_` | Timestamp-based migrations (Flyway-style) |
| `999_grant_all_permissions.sql` | Always runs last — grants |
| `fix_*` | Hotfix migrations (idempotent) |

## Notes

- All migrations are **idempotent** (`IF NOT EXISTS`, `ON CONFLICT DO NOTHING`)  
- `backend/migrations/` and `database/migrations/` are kept in sync  
- Never modify migrations that have already run in production; add a new one instead
