# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Local development (requires Go 1.26+ and PostgreSQL running):**
```bash
# Load env vars (Linux/macOS)
export $(cat .env | xargs)

# Run migration once before first run
go run cmd/migration/main.go

# Run the bot
go run main.go
```

**Docker (full stack):**
```bash
./build.sh                              # Build both Docker images
VERSION=$(cat VERSION) docker-compose up  # Start PostgreSQL + migration + bot
docker-compose down
```

The only tests are in `cmd/migration/` (migration file name parsing and ordering); run them with `go test ./...`.

## Architecture

**Entry point:** `main.go` — loads config, connects to DB, builds a `db.Store`, registers bot commands, launches `inputs.StartScheduler` (weekly/monthly report cron) in a goroutine, then runs a polling loop that dispatches each Telegram update to `inputs.HandleUpdate` in its own goroutine.

**Package layout:**

| Package | Purpose |
|---|---|
| `config/` | Loads and validates 6 required env vars; provides `PgxDSN()` |
| `db/` | `db.Store` wraps `*pgxpool.Pool` and holds all data-access methods: user upsert, category/habit CRUD (soft-delete via `actual`, scoped to owning user), habit tracking + count queries, today/weekly/monthly status queries |
| `inputs/` | Telegram message/callback handling, keyboard builders, in-memory session state for multi-step flows, and the report scheduler |
| `cmd/migration/` | Standalone binary that applies every `migrations/v<N>__<name>.sql` not yet recorded in `schema_migrations` |

**Request flow:** `main.go` update loop → `inputs/handler.go`'s `HandleUpdate` upserts the sender, then dispatches to `handleCommand`, `handleCallback` (by exact match or `strings.HasPrefix` on callback data), or `handleSessionInput` (if the chat has a pending multi-step flow) → handler calls into `db.Store` and sends a message or inline keyboard back.

**Multi-step conversations:** `inputs/session.go` holds a mutex-guarded `map[chatID]session` used by the add/rename category and habit flows — a button tap sets a pending flow, and the next plain-text message from that chat completes it. Delete flows and tracking do **not** use sessions; they're pure button-driven actions.

**Keyboard UI pattern:** `inputs/keyboards.go` builds `tgbotapi.InlineKeyboardMarkup` structs; callback data encodes the action and payload following an `"<entity>:<action>[:<id>]"` convention (e.g. `category:edit:123`, `track:habit:45`). `handler.go` routes `CallbackQuery` updates by matching these prefixes.

**Scheduled reports:** `inputs/scheduler.go`'s `StartScheduler` runs a 1-minute `time.Ticker` and, using the server's local timezone, pushes a weekly report to every user every Monday 08:00 (covering the prior Mon–Sun) and a monthly report on the 1st at 08:00 (covering the prior calendar month). Reports show a per-day tracked count per habit, grouped by category. The "already sent today" guard is in-memory only (no persisted dedup state).

**Daily reminders:** reminders are set per habit. `/reminder` (button-only, no session) lists the user's current reminders, then walks category → habit → hour (06–23) or Off via `reminder:cat:<id>`, `reminder:habit:<id>`, `reminder:set:<habitID>:<hour>` and `reminder:off:<habitID>`, storing the hour in `habit_reminders` (Off deletes the row). On every scheduler tick `inputs/reminder.go`'s `sendDueReminders` finds habits whose hour is the current hour and whose `last_sent_on` is before today, sends each user one message with those of them that have no log today as `track:habit:<id>` buttons (nothing if all are tracked), and sets `last_sent_on`. Changing a habit's hour clears `last_sent_on`. That guard is persisted, so a restart cannot repeat a reminder, and a reminder still goes out if the bot was down at the top of the hour.

**Message length:** `inputs/handler.go`'s `send()` transparently splits any outgoing message over Telegram's 4096-character limit into multiple messages, breaking on line boundaries.

## Database Schema

Four tables in `migrations/v1__init.sql`:
- `users` — Telegram user ID + username, `actual` flag for soft deletes
- `categories` — up to 5 per user (enforced in app logic, not DB); `user_id` is `NOT NULL` with an index
- `habits` — up to 10 per category (enforced in app logic, not DB); `category_id` has an index
- `habit_logs` — one row per tracking event; a habit **can be tracked multiple times per day** (no uniqueness constraint), indexed on `(habit_id, tracked_at)` for the count/report queries

`migrations/v2__reminders.sql` added a per-user `reminders` table; `migrations/v3__habit_reminders.sql` drops it and adds `habit_reminders` — at most one row per habit: `habit_id` (primary key, cascades on delete), `hour` (0–23), and `last_sent_on` (the day the reminder was last handled). A habit with no row has no reminder.

The migration binary (`cmd/migration/`) connects directly with `pgx.Connect()` (not the pool). It keeps a `schema_migrations (version, name, applied_at)` table and applies each `migrations/v<N>__<name>.sql` whose version is not recorded, in numeric order, each in its own transaction together with its `schema_migrations` row. A database that already has the `users` table but an empty `schema_migrations` is baselined at v1 without running it. Schema changes go in a new `v<N>__<name>.sql` file; never edit an applied one.

## Configuration

All config comes from environment variables (see `.env`, which is Git-ignored):

```
TELEGRAM_HABIT_TRACKER_TOKEN
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
```

Missing any variable causes a fatal error at startup. The `config.Config` struct has nested `Telegram` and `Database` sub-structs.

`docker-compose.yml` additionally requires `BACKUP_DIR` in `.env`: the host folder for database dumps. Write it with forward slashes on Windows (`D:/folder`), outside the repo and on a different disk from Docker's data. The bot itself does not read it.

## Implementation Status

All core features are wired up end to end: user upsert, category/habit CRUD (conversational add/rename, picker-based delete), `/trackhabit` (unlimited tracks per day, shown as counts), `/trackpast`, `/untrack`, `/todaystatus`, `/paststatus`, `/reminder` (a daily reminder hour per habit), and automatic weekly/monthly text reports. Not implemented: per-user timezones (schedule and day boundaries use the server's local time) and splitting a single report across multiple Telegram messages beyond the generic length-based split in `send()`.

## Docker

Two images are built by `build.sh`, versioned from the `VERSION` file:
- `habit-tracker-bot:VERSION` — the bot
- `habit-tracker-migration:VERSION` — runs once to apply the schema

`.env` keeps `DB_HOST=localhost` for running outside Docker; `docker-compose.yml` overrides it to `postgres` for the migration and bot services.

In `docker-compose.yml`, the migration service must exit cleanly before the bot starts (`depends_on: migration: condition: service_completed_successfully`).

## Backups

The `backup` service in `docker-compose.yml` runs `scripts/backup.sh` on the same `postgres:17-alpine` image as the database (bump both together; `pg_dump` must match the server's major version). It writes `pg_dump --format=custom` dumps named `habit_tracker-YYYY-MM-DD_HHMM.dump` to `BACKUP_DIR` every night at 03:00, and on start if no dump is under 26 hours old. It keeps the last 14 nightly dumps plus the first dump of each month for 12 months, and its healthcheck fails when the newest dump is older than 26 hours. Dumps are local only and unencrypted.

```bash
# On-demand dump (do this before any upgrade or schema change)
docker-compose exec backup sh /scripts/backup.sh run

# Test restore into a throwaway database, never over the live one
docker-compose exec backup sh -c '
  createdb restore_test &&
  pg_restore --dbname=restore_test --no-owner --exit-on-error /backups/habit_tracker-YYYY-MM-DD_HHMM.dump &&
  psql -d restore_test -c "select count(*) from habit_logs" ;
  dropdb restore_test'

# Real restore (replaces the live database): stop the bot, dump the current state, restore, start
docker-compose stop bot
docker-compose exec backup sh /scripts/backup.sh run
docker-compose exec backup sh -c 'pg_restore --dbname="$PGDATABASE" --clean --if-exists --no-owner --exit-on-error /backups/habit_tracker-YYYY-MM-DD_HHMM.dump'
docker-compose start bot
```

`PREFIX` in the script must not contain a hyphen, because the monthly pruning splits file names on `-`.

The `Dockerfile`'s builder base image (`golang:X-alpine`) must stay at or above the Go version in `go.mod`'s `go` directive, or `go mod download` fails inside the build — bump both together.
