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

## Prerequisite: fix the migration runner

Any schema change will probably break deploys as things stand. `cmd/migration/main.go` always runs `migrations/v1__init.sql`, which uses plain `CREATE TABLE`. With the `postgres_data` volume kept between runs, the migration should fail with "relation already exists" every time after the first `docker-compose up`. Because the bot depends on `service_completed_successfully`, the bot wouldn't start either. This hasn't been confirmed by running it.

Proposed fix:
- Add a `schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMP)` table.
- Have the migration binary apply any `migrations/v*.sql` files not yet recorded, in order, each in its own transaction.
- On an existing database where the v1 tables already exist, record v1 as applied without running it.
- Put the reminders table in `migrations/v2__reminders.sql`.
