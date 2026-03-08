package routes

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-jackson/noah2.0/api/internal/calendar"
	"github.com/noah-jackson/noah2.0/api/internal/db"
	"github.com/noah-jackson/noah2.0/api/internal/llm"
	"github.com/noah-jackson/noah2.0/api/internal/notes"
)

const planningQuery = "weekly priorities goals tasks focus areas"
const topKNotes = 5

// RegisterPlanWeek registers POST /plan/week.
// It fetches upcoming calendar events and semantically relevant notes,
// then asks the LLM to produce a weekly plan.
func RegisterPlanWeek(r *gin.Engine, database *db.DB, llmClient *llm.Client) {
	r.POST("/plan/week", func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. Fetch this week's calendar events from the DB.
		events, err := calendar.FetchWeekEvents(ctx, database)
		if err != nil {
			log.Printf("ERROR fetch week events: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch calendar events"})
			return
		}

		// 2. Embed the planning query and retrieve relevant note chunks.
		embedding, err := llmClient.Embed(ctx, planningQuery)
		if err != nil {
			log.Printf("ERROR embed planning query: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to embed planning query"})
			return
		}

		noteChunks, err := notes.SearchChunks(ctx, database, embedding, topKNotes)
		if err != nil {
			log.Printf("ERROR search note chunks: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search notes"})
			return
		}

		// 3. Build the prompt.
		prompt := buildPlanPrompt(events, noteChunks)

		// 4. Call the LLM.
		plan, err := llmClient.Respond(ctx, prompt)
		if err != nil {
			log.Printf("ERROR llm respond: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to generate plan"})
			return
		}

		log.Printf("INFO generated weekly plan (%d events, %d note chunks)", len(events), len(noteChunks))
		c.JSON(http.StatusOK, gin.H{"plan": plan})
	})
}

func buildPlanPrompt(events []calendar.Event, noteChunks []string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Today is %s.\n\n", time.Now().Format("Monday, January 2, 2006")))

	// Calendar section.
	if len(events) == 0 {
		b.WriteString("I have no calendar events scheduled for the next 7 days.\n\n")
	} else {
		b.WriteString("My calendar events for the next 7 days:\n")
		for _, e := range events {
			if e.IsAllDay {
				b.WriteString(fmt.Sprintf("- %s: %s (all day)\n",
					e.StartTime.Format("Mon Jan 2"), e.Summary))
			} else {
				b.WriteString(fmt.Sprintf("- %s %s–%s: %s\n",
					e.StartTime.Format("Mon Jan 2"),
					e.StartTime.Format("3:04pm"),
					e.EndTime.Format("3:04pm"),
					e.Summary))
			}
		}
		b.WriteString("\n")
	}

	// Notes section.
	if len(noteChunks) > 0 {
		b.WriteString("Relevant context from my notes:\n")
		for i, chunk := range noteChunks {
			b.WriteString(fmt.Sprintf("[Note %d]\n%s\n\n", i+1, chunk))
		}
	}

	b.WriteString("Based on the above, write a concise and practical weekly plan for me. ")
	b.WriteString("Include what to focus on each day, any important events to prepare for, and key priorities.")

	return b.String()
}
