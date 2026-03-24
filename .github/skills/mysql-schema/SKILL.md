---
name: mysql-schema
description: "Write MySQL SQL schemas and migration files. Use when creating tables, altering columns, adding indexes, or writing sql-migrate migration files for the Fiagram project. Use when asked about database schema design or MySQL DDL conventions."
---

# MySQL Schema & Migration Writing

## When to Use
- Creating new database tables
- Altering existing tables (add/drop/modify columns, indexes, constraints)
- Writing new `sql-migrate` migration files
- Reviewing or fixing SQL schema definitions

## Project Migration Setup
- Tool: [rubenv/sql-migrate](https://github.com/rubenv/sql-migrate)
- Dialect: MySQL
- Migration directory: `internal/dao/database/migrations/mysql/`
- Config file: `internal/dao/database/migrations/sql-migrate-config.yaml`
- Migrations are embedded at build time via Go `//go:embed`
- File naming: `<YYYYMMDDHHMMSS>-<description>.sql` (e.g., `20251228180902-init.sql`)

## Migration File Structure
Every migration file MUST have both `Up` and `Down` sections:

```sql
-- +migrate Up
<create/alter statements>

-- +migrate Down
<drop/revert statements in reverse dependency order>
```

## Schema Conventions (from existing codebase)

### Data Types
- Primary keys: `BIGINT UNSIGNED AUTO_INCREMENT` for entity tables, `INT UNSIGNED` for lookup/enum tables
- Foreign keys: match the referenced column type exactly (e.g., `BIGINT UNSIGNED` → `BIGINT UNSIGNED`)
- Strings: `VARCHAR(n)` with appropriate length — `VARCHAR(255)` general, `VARCHAR(500)` for URLs, `VARCHAR(128)` for hashes, `VARCHAR(15)` for phone numbers, `VARCHAR(20)` for short names/enums
- Timestamps: `TIMESTAMP NOT NULL`

### Column Naming
- Use `snake_case` for all column and table names
- Foreign key columns: prefix with `of_` (e.g., `of_account_id`, `of_role_id`)
- Table names: plural nouns (e.g., `accounts`, `account_passwords`, `chatbot_webhooks`)
- Lookup/enum tables: singular concept + `_role`, `_type`, `_status` (e.g., `account_role`)

### Timestamps
Every mutable entity table MUST include:
```sql
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
```

### Keys & Constraints
- Always declare `PRIMARY KEY (id)` or `PRIMARY KEY (of_<parent>_id)` for 1:1 tables
- Foreign keys: `FOREIGN KEY (of_<parent>_id) REFERENCES <parent_table>(id)`
- Add `UNIQUE` constraints for natural keys (e.g., `username`, `email`)
- Composite unique constraints where needed: `UNIQUE (of_account_id, name)`

### Index Naming
- Regular index: `idx_<table>_<column>` (e.g., `idx_accounts_email`)
- Composite index: `idx_<table>_<col1>_<col2>` (e.g., `idx_chatbot_webhooks_of_account_id_name`)
- Unique index: `udx_<table>_<column>` (e.g., `udx_accounts_username`)
- Foreign key index: `fk_<table>_<column>` (e.g., `fk_accounts_of_role_id`)

### Table Patterns
| Pattern | PK | FK | Example |
|---------|----|----|---------|
| Entity table | `BIGINT UNSIGNED AUTO_INCREMENT` | — | `accounts` |
| Lookup/enum table | `INT UNSIGNED` (no auto-increment) | — | `account_role` |
| 1:1 extension table | FK as PK (`BIGINT UNSIGNED`) | references parent | `account_passwords` |
| 1:N child table | `BIGINT UNSIGNED AUTO_INCREMENT` | references parent | `chatbot_webhooks` |

### Seed Data
- Enum/lookup tables: include `INSERT` statements in the same migration
- Use explicit `id` values for enums (e.g., `0='none'`, `1='admin'`, `2='member'`)

## Down Migration Rules
- Drop tables in **reverse dependency order** (children before parents)
- Use `DROP TABLE IF EXISTS` for safety
- For `ALTER TABLE` changes, write the exact reverse operation
- For seed data `INSERT`, use `DELETE` in the down migration

## Procedure
1. Read the existing migrations in `internal/dao/database/migrations/mysql/` to understand current schema state
2. Write the migration SQL following all conventions above
3. Create the file with the correct timestamp-based name
4. Verify both `Up` and `Down` sections are complete and reversible
5. Run `make migrate-up-dev` to apply and verify the migration succeeds

## Makefile Commands
- `make migrate-up-dev` — apply pending migrations
- `make migrate-down-dev` — rollback last migration
- `make migrate-new <name>` — scaffold a new migration file
- `make migrate-status` — show migration status
