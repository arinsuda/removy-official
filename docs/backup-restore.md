# POPCORN Database Backup & Restore

POPCORN runs on **Neon Postgres** (PostgreSQL 16). Because the Neon free plan limits point-in-time recovery (PITR) history to roughly **6 hours**, `.github/workflows/db-backup.yml` runs a daily encrypted logical backup (`pg_dump` piped into `age`) and retains artifacts for **30 days** (`BACKUP_RETENTION_DAYS`).

The repository may be public: **workflow logs and artifacts must be treated as visible to others**. Backups are always encrypted on the runner using an asymmetric `age` public key before any artifact is written to GitHub storage, and the private decryption key never leaves your offline vault.

---

## 1. One-Time Setup

### 1.1 Generate the offline `age` keypair

Install [`age`](https://github.com/FiloSottile/age) locally, then generate a dedicated keypair:

```bash
age-keygen -o popcorn-backup.key
chmod 600 popcorn-backup.key
```

The command prints the public key to stderr (and writes it as a comment inside `popcorn-backup.key`), looking like:

```text
# public key: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

1. Save the entire `popcorn-backup.key` file (the `AGE-SECRET-KEY-1...` line) in your **password manager or offline vault**, then remove loose copies from disk.
   - **Warning:** Never commit `popcorn-backup.key` to Git and never put it in GitHub Secrets or Variables.
   - **Warning:** Losing `popcorn-backup.key` makes every existing `.dump.age` backup permanently unreadable.
2. Add the **public key** (`age1...`) as the GitHub repository secret **`BACKUP_AGE_RECIPIENT`**.

### 1.2 Create a read-only Neon role for backups

Connect to your Neon production database as the database/migration owner role (typically `neondb_owner`, or whichever role creates tables during server startup) and run:

```sql
CREATE ROLE popcorn_backup WITH LOGIN PASSWORD '<STRONG_GENERATED_PASSWORD>';

GRANT CONNECT ON DATABASE neondb TO popcorn_backup;
GRANT USAGE ON SCHEMA public TO popcorn_backup;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO popcorn_backup;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO popcorn_backup;

-- Ensure future tables and sequences added by migrations are readable too.
-- Run this as the migration owner role (e.g. neondb_owner), or specify FOR ROLE <owner_role>.
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO popcorn_backup;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON SEQUENCES TO popcorn_backup;
```

> [!IMPORTANT]
> `GRANT SELECT ON ALL SEQUENCES` and `ALTER DEFAULT PRIVILEGES ... SEQUENCES` are required because `pg_dump` reads `last_value` and `is_called` from sequences.

### 1.3 Use the DIRECT (non-pooler) Neon endpoint

Neon exposes two hostnames per branch:
- **Pooled (`-pooler` in hostname):** runs PgBouncer in transaction pooling mode. **Do not use this for `pg_dump`** — `pg_dump` relies on `REPEATABLE READ` snapshot transactions and session settings that fail or behave inconsistently through a transaction pooler.
- **Direct (no `-pooler` in hostname):** direct PostgreSQL 16 endpoint. **Use this for `BACKUP_DATABASE_URL`.**

Construct the connection string with `sslmode=require` and save it as the GitHub repository secret **`BACKUP_DATABASE_URL`**:

```text
postgres://popcorn_backup:<STRONG_GENERATED_PASSWORD>@<direct-neon-host>/neondb?sslmode=require
```

### 1.4 Verify GitHub Actions configuration

In GitHub repository **Settings -> Secrets and variables -> Actions -> Secrets**, ensure both secrets exist:
- `BACKUP_DATABASE_URL` — direct, read-only Neon connection URI with `sslmode=require`
- `BACKUP_AGE_RECIPIENT` — `age1...` public recipient key

Trigger the **Daily Encrypted Database Backup** workflow manually once via `workflow_dispatch` to confirm it completes green.

---

## 2. How Automated Verification Works (and Why Drills Are Manual)

Each run of `.github/workflows/db-backup.yml`:
1. Installs PostgreSQL 16 client tools (`postgresql-client-16`) and fails immediately if either the client or remote server major version is not `16`.
2. Runs with `set -euo pipefail`, streaming `pg_dump --format=custom --no-owner --no-privileges` directly into `age -r "$BACKUP_AGE_RECIPIENT"`. No plaintext `.dump` or `.sql` file ever touches runner disk.
3. Checks that `pg_dump` exited `0`, the encrypted file size exceeds `BACKUP_MIN_SIZE_BYTES`, the `age-encryption.org/v1` header is present, and no plaintext `PGDMP` header exists.
4. Uploads `popcorn-db-<UTC timestamp>.dump.age` with retention set from `BACKUP_RETENTION_DAYS` (`30`).
5. Runs a `verify` job with a `postgres:16` service container, downloads the artifact, and re-verifies integrity without decrypting.

Because the `age` private key never exists on GitHub, **CI cannot decrypt or restore the dump — a full restore drill is manual** and should be run periodically (for example, monthly or before major migrations).

---

## 3. Recovery Decision Table

| Scenario / Discovery Delay | Recommended Recovery Path | Notes |
|---|---|---|
| Incident discovered **< 6 hours** ago | **Neon point-in-time branch** | Use Neon's branch/PITR UI to branch to the exact timestamp before the mistake, inspect data, then promote or copy rows. |
| Incident discovered **> 6 hours** ago (up to 30 days) | **Daily encrypted backup (`.dump.age`)** | Download the latest healthy artifact before the incident and restore with `scripts/db-restore.sh` to a fresh Neon branch or local Postgres 16 first. |
| Partial data loss (e.g., accidental delete of specific rows) | **Restore to scratch target, then selectively copy** | Restore into a temporary Neon branch or local Docker Postgres, verify rows, and backfill only the missing records into production so newer writes aren't lost. |

---

## 4. Manual Restore Drill

Prerequisites locally:
- `age`
- PostgreSQL 16 client tools (`pg_restore` and `psql` major version `>= 16`; on macOS with Homebrew: `brew install postgresql@16 age`, or run inside a `postgres:16` Docker container)
- `gh` CLI (or GitHub Actions web UI) to download the artifact

### Step 1: Download the encrypted artifact

From the GitHub Actions run page, download `popcorn-db-<UTC timestamp>.dump.age` (or via `gh`):

```bash
gh run download <RUN_ID> -n popcorn-db-<UTC_TIMESTAMP>.dump.age
```

### Step 2: Prepare a clean restore target

Choose either **Option A (Local Docker Postgres 16)** or **Option B (Temporary Neon branch)**:

- **Option A — Local Docker Postgres 16 (safest for routine drills):**

  ```bash
  docker run -d --name popcorn_restore_drill \
    -e POSTGRES_USER=restore \
    -e POSTGRES_PASSWORD=restore \
    -e POSTGRES_DB=popcorn_restore \
    -p 55432:5432 \
    postgres:16-alpine
  ```

- **Option B — Isolated Neon branch:**
  1. In the Neon console, create a temporary branch (e.g., `restore-drill-YYYYMMDD`).
  2. Copy its **direct** (non-pooler) connection string with `sslmode=require`.

### Step 3: Retrieve the private key and run `scripts/db-restore.sh`

Copy your private key from your password manager into a temporary file (`chmod 600 popcorn-backup.key`), then run:

```bash
# Against local Docker Postgres 16:
./scripts/db-restore.sh \
  popcorn-db-<UTC_TIMESTAMP>.dump.age \
  popcorn-backup.key \
  "postgres://restore:restore@localhost:55432/popcorn_restore"

# Against a Neon branch (scripts/db-restore.sh blocks *.neon.tech unless --i-know-this-is-prod is passed):
./scripts/db-restore.sh --i-know-this-is-prod \
  popcorn-db-<UTC_TIMESTAMP>.dump.age \
  popcorn-backup.key \
  "postgres://<user>:<password>@<branch-direct-host>/neondb?sslmode=require"
```

### Step 4: Verify counts and clean up

`scripts/db-restore.sh` prints `total_rows` and `active_rows` (`deleted_at IS NULL`) for `users` and `reviews` after `pg_restore` finishes.

1. Compare those counts against production (or expected daily totals) and spot-check recent rows in `psql`.
2. Delete the local copy of `popcorn-backup.key` and the `.dump.age` artifact when finished:

   ```bash
   rm -f popcorn-backup.key popcorn-db-<UTC_TIMESTAMP>.dump.age
   docker rm -f popcorn_restore_drill   # if using Option A
   ```
   If using Option B, delete the temporary Neon branch in the Neon console.

---

## 5. PDPA & Erasure Retention Note

POPCORN implements PDPA account erasure (`account_erasure`, including the 30-day self-erasure / admin soft-delete grace period and permanent hard deletion of user records, reviews, comments, likes, lists, follows, and notifications).

- **30-day backup retention:** Encrypted daily backup artifacts are retained on GitHub Actions for **30 days** (`BACKUP_RETENTION_DAYS`). Consequently, personal data of a user erased via `account_erasure` remains inside previously created `.dump.age` backup artifacts for up to **30 days** after erasure, until those artifacts expire and are purged automatically.
- **Post-restore hygiene:** If you ever restore production (or a long-lived staging database) from a backup taken before a user's erasure completed, re-apply any erasure actions recorded after that backup's timestamp (check `admin_audit_logs` for `USER_ERASURE_REQUESTED`, `USER_ERASED`, and `USER_ERASED_SCHEDULED` events).
