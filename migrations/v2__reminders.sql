CREATE TABLE reminders (
    user_id      BIGINT PRIMARY KEY REFERENCES users(id),
    hour         SMALLINT NOT NULL CHECK (hour BETWEEN 0 AND 23),
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    last_sent_on DATE
);
