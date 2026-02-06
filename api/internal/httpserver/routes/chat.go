package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noah-jackson/noah2.0/api/internal/config"
	"github.com/noah-jackson/noah2.0/api/internal/llm"
)

type chatRequest struct {
	Message string `json:"message"`
}

func RegisterChat(r *gin.Engine, cfg config.Config, client *llm.Client) {
	r.POST("/chat", func(c *gin.Context) {
		if cfg.OpenAIAPIKey == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "OPENAI_API_KEY is not set",
			})
			return
		}

		var req chatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid JSON body",
			})
			return
		}

		req.Message = strings.TrimSpace(req.Message)
		if req.Message == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "message is required",
			})
			return
		}

		reply, err := client.Respond(c.Request.Context(), req.Message)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"reply": reply,
		})
	})
}
