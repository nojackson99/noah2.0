-- Run with:
--   psql $DATABASE_URL -f api/db/migrations/001_calendar_events.sql

CREATE TABLE IF NOT EXISTS calendar_events (
    google_event_id  TEXT        PRIMARY KEY,
    calendar_id      TEXT        NOT NULL DEFAULT 'primary',
    summary          TEXT,
    description      TEXT,
    location         TEXT,
    start_time       TIMESTAMPTZ NOT NULL,
    end_time         TIMESTAMPTZ NOT NULL,
    is_all_day       BOOLEAN     NOT NULL DEFAULT FALSE,
    status           TEXT,
    html_link        TEXT,
    ingested_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_start_time
    ON calendar_events (start_time);
