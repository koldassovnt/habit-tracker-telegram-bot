-- Reminders are set per habit; the per-user reminder from v2 is replaced.
DROP TABLE reminders;

CREATE TABLE habit_reminders (
    habit_id     BIGINT PRIMARY KEY REFERENCES habits(id) ON DELETE CASCADE,
    hour         SMALLINT NOT NULL CHECK (hour BETWEEN 0 AND 23),
    last_sent_on DATE
);

CREATE INDEX ON habit_reminders (hour);
