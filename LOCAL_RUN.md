# Running Locally Without Docker

## Prerequisites

- Go 1.24+
- A running PostgreSQL instance

If you don't have PostgreSQL installed locally, you can spin up a quick instance with Docker:

```bash
docker run --name habit-postgres \
  -e POSTGRES_USER=habit_user \
  -e POSTGRES_PASSWORD=habit_password \
  -e POSTGRES_DB=habit_tracker \
  -p 5432:5432 \
  -d postgres:17-alpine
```

---

## Step 1 — Set Environment Variables

Load them from your `.env` file:

```bash
export $(cat .env | xargs)
```

Or export them manually:

```bash
export TELEGRAM_HABIT_TRACKER_TOKEN=your_token_here
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=habit_user
export DB_PASSWORD=habit_password
export DB_NAME=habit_tracker
```

---

## Step 2 — Run Migration

Run this once to create the database tables:

```bash
go run cmd/migration/main.go
```

You should see:

```
Migration done
```

Only needs to be run once. Skip it on subsequent starts unless the schema has changed.

---

## Step 3 — Run the Bot

```bash
go run main.go
```

You should see:

```
Database connected
Bot @your_bot_name is running...
```

---

## Stopping the Bot

Press `Ctrl+C` — the bot will shut down gracefully.
