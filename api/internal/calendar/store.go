package calendar

import (
	"context"
	"fmt"

	"github.com/noah-jackson/noah2.0/api/internal/db"
)

// UpsertEvents inserts or updates a slice of events in the database.
// Keyed on google_event_id, so running ingest multiple times is idempotent.
func UpsertEvents(ctx context.Context, database *db.DB, events []Event) (int64, error) {
	const q = `
		INSERT INTO calendar_events (
			google_event_id,
			calendar_id,
			summary,
			description,
			location,
			start_time,
			end_time,
			is_all_day,
			status,
			html_link,
			ingested_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
		)
		ON CONFLICT (google_event_id) DO UPDATE SET
			summary     = EXCLUDED.summary,
			description = EXCLUDED.description,
			location    = EXCLUDED.location,
			start_time  = EXCLUDED.start_time,
			end_time    = EXCLUDED.end_time,
			is_all_day  = EXCLUDED.is_all_day,
			status      = EXCLUDED.status,
			html_link   = EXCLUDED.html_link,
			updated_at  = NOW()
	`

	var upserted int64
	for _, e := range events {
		tag, err := database.Pool.Exec(ctx, q,
			e.GoogleEventID,
			e.CalendarID,
			nullableString(e.Summary),
			nullableString(e.Description),
			nullableString(e.Location),
			e.StartTime,
			e.EndTime,
			e.IsAllDay,
			nullableString(e.Status),
			nullableString(e.HTMLLink),
		)
		if err != nil {
			return upserted, fmt.Errorf("upsert event %q: %w", e.GoogleEventID, err)
		}
		upserted += tag.RowsAffected()
	}

	return upserted, nil
}

// nullableString converts an empty string to nil so pgx stores it as SQL NULL.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
