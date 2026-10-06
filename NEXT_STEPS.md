# Next Steps

Version 1.1.0 (backups, the versioned migration runner and `/reminder`) was deployed on 2026-10-06. On the same day the bot moved from two hand-started `docker run` containers onto `docker-compose`, with the database restored from a dump into the `postgres_data` volume. Backups, the test restore and the migration were checked then.

## Left to do

1. **Check reminders in Telegram.** Send `/reminder` and pick the current hour. Within a minute the bot should list only the habits not tracked today; tapping one tracks it. Restart the bot in the same hour: no second reminder. `/reminder` → Off stops them.
2. **Remove the old containers** once the compose stack has run well for a few days. They are stopped and still hold the pre-move database as a fallback:
   ```bash
   docker rm -v habit-tracker-bot habit-postgres
   ```
   The same data is also in `pre-compose-move-final-2026-10-06.dump` in `BACKUP_DIR`. To fall back before then: `docker-compose down` (without `-v`), then `docker start habit-postgres habit-tracker-bot`.

## Ideas not planned yet

- Per-user timezones. Reminders and reports use server local time (Asia/Almaty).
- Reminder times per habit (e.g. "Drink water" at 10, 14 and 18).
- An off-site or encrypted copy of the backups. They are local only, which survives a disk failure or a Docker reset, not theft or fire.
- Saving the weekly/monthly report "already sent" guard in the database, as reminders do with `last_sent_on`.
