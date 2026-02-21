package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/noah-jackson/noah2.0/api/internal/calendar"
	"github.com/noah-jackson/noah2.0/api/internal/db"
)

// RegisterIngest registers POST /ingest/calendar.
// It fetches upcoming events from Google Calendar and upserts them into Postgres.
func RegisterIngest(r *gin.Engine, calClient *calendar.Client, database *db.DB) {
	r.POST("/ingest/calendar", func(c *gin.Context) {
		events, err := calClient.FetchUpcoming(c.Request.Context())
		if err != nil {
			log.Printf("ERROR fetch calendar events: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "failed to fetch events from Google Calendar",
			})
			return
		}

		count, err := calendar.UpsertEvents(c.Request.Context(), database, events)
		if err != nil {
			log.Printf("ERROR upsert calendar events: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to store events",
			})
			return
		}

		log.Printf("INFO ingested %d calendar events (%d fetched)", count, len(events))
		c.JSON(http.StatusOK, gin.H{
			"fetched":  len(events),
			"upserted": count,
		})
	})
}
