# Database Migrations

This directory contains SQL scripts for schema changes and seeds.

## Execution Order
1. `001_create_schema.sql` – creates enums, tables, triggers, and audit columns.
2. `seed/001_seed.sql` – inserts seed data aligned with the new schema.

Run each script in order using your preferred PostgreSQL tool, e.g.:

```sh
psql -f migrations/001_create_schema.sql
psql -f migrations/seed/001_seed.sql
```

## Rollback
To revert the changes, execute the SQL statements in the `Down` section of `001_create_schema.sql` (reverse order of creation). If you are using a migration tool, run the corresponding down migration.
