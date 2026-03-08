package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/noah-jackson/noah2.0/api/internal/db"
)

// FetchWeekEvents returns calendar events from the DB for the next 7 days.
func FetchWeekEvents(ctx context.Context, database *db.DB) ([]Event, error) {
	const q = `
		SELECT
			google_event_id,
			COALESCE(summary, ''),
			COALESCE(description, ''),
			COALESCE(location, ''),
			start_time,
			end_time,
			is_all_day,
			COALESCE(status, '')
		FROM calendar_events
		WHERE start_time >= $1
		  AND start_time <= $2
		ORDER BY start_time ASC
	`

	now := time.Now()
	start := now.Truncate(24 * time.Hour)
	end := start.AddDate(0, 0, 7)

	rows, err := database.Pool.Query(ctx, q, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetch week events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.GoogleEventID,
			&e.Summary,
			&e.Description,
			&e.Location,
			&e.StartTime,
			&e.EndTime,
			&e.IsAllDay,
			&e.Status,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}

	return events, rows.Err()
}
