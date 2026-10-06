# Next Steps

Backups, the versioned migration runner and daily reminders (`/reminder`) are implemented in version 1.1.0. They have been built and unit-tested, but not yet run against the real database. Roll them out in this order.

## Rollout checklist

1. **Backups first.** Add `BACKUP_DIR=D:/some-folder` to `.env` (forward slashes, outside the repo, on a different disk from Docker's data) and create the folder. Then start only the backup service, which does not run the migration:
   ```bash
   VERSION=$(cat VERSION) docker-compose up -d backup
   docker-compose logs backup      # expect "wrote habit_tracker-....dump"
   docker-compose ps backup        # expect "healthy" after a couple of minutes
   ```
2. **Test the restore** into a throwaway database:
   ```bash
   docker-compose exec backup sh -c '
     createdb restore_test &&
     pg_restore --dbname=restore_test --no-owner --exit-on-error /backups/habit_tracker-YYYY-MM-DD_HHMM.dump &&
     psql -d restore_test -c "select count(*) from habit_logs" ;
     dropdb restore_test'
   ```
3. **Deploy 1.1.0.**
   ```bash
   ./build.sh
   VERSION=$(cat VERSION) docker-compose up -d
   docker-compose logs migration
   ```
   On the existing database the migration log should say it recorded `v1__init.sql` as already applied, then `Applied v2__reminders.sql`. A second `docker-compose up` should log `Nothing to apply`.
4. **Check reminders in Telegram.** Send `/reminder` and pick the current hour. Within a minute the bot should list only the habits not tracked today; tapping one tracks it. Restart the bot in the same hour: no second reminder. `/reminder` → Off stops them.

## Ideas not planned yet

- Per-user timezones. Reminders and reports use server local time (Asia/Almaty).
- Reminder times per habit (e.g. "Drink water" at 10, 14 and 18).
- An off-site or encrypted copy of the backups. They are local only, which survives a disk failure or a Docker reset, not theft or fire.
- Saving the weekly/monthly report "already sent" guard in the database, as reminders do with `last_sent_on`.
