# Next Steps

Nothing is pending. On 2026-10-06 the bot moved from two hand-started `docker run` containers onto `docker-compose`, and versions 1.1.0 (nightly backups, the versioned migration runner) and 1.2.0 (per-habit `/reminder`) were deployed and checked. The old containers were removed; the pre-move database is kept as `pre-compose-move-final-2026-10-06.dump` in `BACKUP_DIR`.

## Ideas not planned yet

- Per-user timezones. Reminders and reports use server local time (Asia/Almaty).
- Several reminder times for one habit (e.g. "Drink water" at 10, 14 and 18). A habit has at most one reminder hour.
- An off-site or encrypted copy of the backups. They are local only, which survives a disk failure or a Docker reset, not theft or fire.
- Saving the weekly/monthly report "already sent" guard in the database, as reminders do with `last_sent_on`.
