# Next Steps

- Per-user timezones. Reminders and reports use server local time (Asia/Almaty).
- Several reminder times for one habit (e.g. "Drink water" at 10, 14 and 18). A habit has at most one reminder hour.
- An off-site or encrypted copy of the backups. They are local only, which survives a disk failure or a Docker reset, not theft or fire.
- Saving the weekly/monthly report "already sent" guard in the database, as reminders do with `last_sent_on`.
