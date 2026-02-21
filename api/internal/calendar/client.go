package calendar

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googlecalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// How far back and forward to fetch events.
const (
	PastDays   = 7
	FutureDays = 30
)

// Client wraps the Google Calendar API service.
type Client struct {
	svc *googlecalendar.Service
}

// NewClient creates a Calendar API client from stored OAuth credentials.
// It uses the refresh token to obtain access tokens automatically.
func NewClient(ctx context.Context, clientID, clientSecret, refreshToken string) (*Client, error) {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{googlecalendar.CalendarReadonlyScope},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// TokenSource will transparently refresh the access token when it expires.
	tokenSource := cfg.TokenSource(ctx, token)

	svc, err := googlecalendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}

	return &Client{svc: svc}, nil
}

// Event is a simplified representation of a Google Calendar event.
type Event struct {
	GoogleEventID string
	CalendarID    string
	Summary       string
	Description   string
	Location      string
	StartTime     time.Time
	EndTime       time.Time
	IsAllDay      bool
	Status        string
	HTMLLink      string
}

// FetchUpcoming retrieves events from the primary calendar for a window
// of PastDays before now through FutureDays after now.
// NOTE: MaxResults is capped at 250. Pagination is not implemented yet.
func (c *Client) FetchUpcoming(ctx context.Context) ([]Event, error) {
	now := time.Now()
	timeMin := now.AddDate(0, 0, -PastDays).Format(time.RFC3339)
	timeMax := now.AddDate(0, 0, FutureDays).Format(time.RFC3339)

	result, err := c.svc.Events.List("primary").
		Context(ctx).
		TimeMin(timeMin).
		TimeMax(timeMax).
		SingleEvents(true). // expand recurring events into individual instances
		OrderBy("startTime").
		MaxResults(250).
		Do()
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}

	events := make([]Event, 0, len(result.Items))
	for _, item := range result.Items {
		e, err := parseEvent(item)
		if err != nil {
			// Log and skip malformed events rather than failing the whole ingest.
			log.Printf("WARN skip malformed event %q: %v", item.Id, err)
			continue
		}
		events = append(events, e)
	}

	return events, nil
}

// parseEvent converts a Google API event object into the local Event type.
func parseEvent(item *googlecalendar.Event) (Event, error) {
	e := Event{
		GoogleEventID: item.Id,
		CalendarID:    "primary",
		Summary:       item.Summary,
		Description:   item.Description,
		Location:      item.Location,
		Status:        item.Status,
		HTMLLink:      item.HtmlLink,
	}

	// Google distinguishes all-day events (Date only) from timed events (DateTime).
	if item.Start.DateTime != "" {
		t, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			return Event{}, fmt.Errorf("parse start datetime %q: %w", item.Start.DateTime, err)
		}
		e.StartTime = t
	} else {
		// All-day event: parse the date and treat as midnight UTC.
		t, err := time.Parse("2006-01-02", item.Start.Date)
		if err != nil {
			return Event{}, fmt.Errorf("parse start date %q: %w", item.Start.Date, err)
		}
		e.StartTime = t
		e.IsAllDay = true
	}

	if item.End.DateTime != "" {
		t, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			return Event{}, fmt.Errorf("parse end datetime %q: %w", item.End.DateTime, err)
		}
		e.EndTime = t
	} else {
		t, err := time.Parse("2006-01-02", item.End.Date)
		if err != nil {
			return Event{}, fmt.Errorf("parse end date %q: %w", item.End.Date, err)
		}
		e.EndTime = t
	}

	return e, nil
}
