# Docker Build & Versioning

## Version Management

The project version is stored in a `VERSION` file at the root of the project:

```
1.0.0
```

Bump this manually before each build.

---

## Project Images

The project builds two Docker images:

| Image | Description |
|---|---|
| `habit-tracker-bot` | The Telegram bot |
| `habit-tracker-migration` | Standalone DB migration script |

---

## Building Images

Use the `build.sh` script to build both images with the current version:

```bash
./build.sh
```

This will tag both images with the version from the `VERSION` file. For example, if `VERSION` contains `1.0.0`:

```
habit-tracker-bot:1.0.0
habit-tracker-migration:1.0.0
```

Make sure the script is executable:

```bash
chmod +x build.sh
```

---

## Running with Docker Compose

The `docker-compose.yml` uses the `VERSION` env var to pick the correct image tags.

Run the full stack:

```bash
VERSION=$(cat VERSION) docker-compose up
```

Run in detached mode:

```bash
VERSION=$(cat VERSION) docker-compose up -d
```

Stop the stack:

```bash
docker-compose down
```

Stop and remove the database volume:

```bash
docker-compose down -v
```

---

## Services

### postgres
Runs a PostgreSQL 17 database. Data is persisted in a named Docker volume `postgres_data` so it survives container restarts.

### migration
Runs once on startup and applies the SQL migration from `migrations/v1_init.sql`. The bot will not start until this completes successfully.

### bot
The Telegram bot. Starts only after the migration service completes successfully.

---

## Environment Variables

All environment variables are loaded from the `.env` file. Copy the example and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description |
|---|---|
| `TELEGRAM_HABIT_TRACKER_TOKEN` | Your Telegram bot token |
| `DB_HOST` | PostgreSQL host |
| `DB_PORT` | PostgreSQL port |
| `DB_USER` | PostgreSQL user |
| `DB_PASSWORD` | PostgreSQL password |
| `DB_NAME` | PostgreSQL database name |

> **Never commit `.env` to git.** It is listed in `.gitignore`.

---

## Running Migration Manually

To run the migration script independently outside of Docker Compose:

```bash
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=habit_user \
DB_PASSWORD=habit_password \
DB_NAME=habit_tracker \
go run cmd/migration/main.go
```
