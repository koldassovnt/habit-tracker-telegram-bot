# Next Steps

## Daily habit reminders

A daily check-in at a time each user picks. It lists only the habits not yet tracked today, with one-tap buttons to track them.

### Behavior

- `/reminder` shows an inline keyboard of hours (06:00–23:00) plus an **Off** button. It is button-only, like the delete flows, so it needs no session.
- At the chosen hour the bot sends "⏰ Not tracked yet today:" with one `track:habit:<id>` button per untracked habit. Tapping a button goes through the existing track handler in `inputs/handler.go`.
- If every habit is already tracked that day, no message is sent.
- Times use server local time (Asia/Almaty), the same as the weekly and monthly reports. Per-user timezones are out of scope.

### Implementation sketch

- **Schema:** new `reminders` table:
  ```sql
  CREATE TABLE reminders (
      user_id      BIGINT PRIMARY KEY REFERENCES users(id),
      hour         SMALLINT NOT NULL CHECK (hour BETWEEN 0 AND 23),
      enabled      BOOLEAN NOT NULL DEFAULT TRUE,
      last_sent_on DATE
  );
  ```
  `last_sent_on` is saved in the database, so a bot restart can't send a reminder twice or skip one. The report scheduler keeps its "already sent" guard in memory, so it doesn't have this protection.
- **db:** add `SetReminder(userID, hour)`, `DisableReminder(userID)`, `DueReminders(now)` (users with `enabled`, `hour = now.Hour()`, `last_sent_on < today OR NULL`), `MarkReminderSent(userID, date)`, and a query for habits with no `habit_logs` row for today.
- **Scheduler:** in the existing 1-minute ticker in `inputs/scheduler.go`, call `DueReminders`, send each message, then mark it sent.
- **Callbacks:** `reminder:set:<hour>` and `reminder:off`, following the `<entity>:<action>[:<id>]` convention.
- Register `/reminder` with the other bot commands in `main.go`.

### Options considered

- **Reminder times per habit** (e.g. "Drink water" at 10, 14 and 18): more flexible, but needs much more UI for choosing a habit and managing a list of times.
- **A fixed evening nudge for everyone** (e.g. 21:00, no settings): the simplest, but users can't change or turn it off.

## PostgreSQL backups

The database lives only in the `postgres_data` Docker volume, so losing the volume or the host loses every user's habits and logs. Add automated backups using the same approach already set up in the `family-finance-crm` project, so both projects are backed up and restored the same way.

### How it works there

- A `backup` sidecar service in the compose file runs `pg_dump --format=custom` every night at 03:00 (Asia/Almaty) using busybox `crond`. It uses the same `postgres:17-alpine` image as the database, so `pg_dump` always matches the server's major version and no extra image is needed.
- Its entrypoint is a POSIX sh script, `db/backup.sh`, bind-mounted read-only, with three modes: `schedule` (dump now if nothing is fresh, then `exec crond -f`), `run` (one dump plus prune, also used on demand) and `healthcheck`.
- Dumps go to a host folder bind-mounted at `/backups` from `${BACKUP_DIR}`, which sits on a different physical disk from Docker's data. There is no off-site copy and the dumps are not encrypted: this survives a disk failure or a Docker reset, not theft or fire.
- Retention: the last 14 nightly dumps plus the first dump of each month for 12 months. Pruning runs after each successful dump.
- The container reports unhealthy when the newest dump is older than 26 hours. On start it takes a dump straight away if none is that fresh.

### What to do here

- Copy `db/backup.sh` from `family-finance-crm` and set `PREFIX=habit_tracker`. The prefix must not contain a hyphen, because the monthly pruning splits file names on `-`.
- Add the `backup` service to `docker-compose.yml`:
  ```yaml
  backup:
    image: postgres:17-alpine
    depends_on:
      postgres:
        condition: service_healthy
    entrypoint: ["sh", "/scripts/backup.sh"]
    command: ["schedule"]
    environment:
      TZ: Asia/Almaty
      PGHOST: postgres
      PGDATABASE: ${DB_NAME}
      PGUSER: ${DB_USER}
      PGPASSWORD: ${DB_PASSWORD}
    volumes:
      - ${BACKUP_DIR:?set BACKUP_DIR in .env}:/backups
      - ./db/backup.sh:/scripts/backup.sh:ro
    healthcheck:
      test: ["CMD", "sh", "/scripts/backup.sh", "healthcheck"]
      interval: 5m
      timeout: 10s
      start_period: 2m
    restart: unless-stopped
    networks:
      - habit-tracker-bot-network
  ```
  The source project puts this service in an `app` compose profile; this stack has no profiles, so it is left out.
- Add `BACKUP_DIR` to `.env`, written with forward slashes on Windows (`D:/folder`), outside the repo and on a different disk from Docker's data. Document it in `CLAUDE.md` with the other variables.
- Do one test restore into a throwaway database after setup, and again after any Postgres major upgrade:
  ```bash
  docker-compose exec backup sh -c '
    createdb restore_test &&
    pg_restore --dbname=restore_test --no-owner --exit-on-error /backups/habit_tracker-YYYY-MM-DD_HHMM.dump &&
    psql -d restore_test -c "select count(*) from habit_logs" ;
    dropdb restore_test'
  ```
- Document the real restore: stop the bot, take one more dump with `docker-compose exec backup sh /scripts/backup.sh run`, then `pg_restore --dbname=$DB_NAME --clean --if-exists --no-owner --exit-on-error <dump>`, and start the bot again.

### Gotchas already solved in the script

- busybox `crond` does not pass the container environment to jobs, so the entrypoint writes the `PG*` and `TZ` exports to `/etc/backup.env` and the cron line sources it.
- Cron output is redirected to `/proc/1/fd/1`; otherwise it never reaches `docker logs`.
- Each dump is written to a hidden `.name.partial` file and renamed on success, so a half-written file never matches the `*.dump` glob that pruning and the healthcheck use.
- The backup image must stay on the same Postgres major version as the database image; bump both together.

## Prerequisite: fix the migration runner

Any schema change will probably break deploys as things stand. `cmd/migration/main.go` always runs `migrations/v1__init.sql`, which uses plain `CREATE TABLE`. With the `postgres_data` volume kept between runs, the migration should fail with "relation already exists" every time after the first `docker-compose up`. Because the bot depends on `service_completed_successfully`, the bot wouldn't start either. This hasn't been confirmed by running it.

Proposed fix:
- Add a `schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMP)` table.
- Have the migration binary apply any `migrations/v*.sql` files not yet recorded, in order, each in its own transaction.
- On an existing database where the v1 tables already exist, record v1 as applied without running it.
- Put the reminders table in `migrations/v2__reminders.sql`.
