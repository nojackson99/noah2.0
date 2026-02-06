package httpserver

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
}

func New(env string) *Server {
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())

	return &Server{Engine: r}
}

func (s *Server) Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("listening on %s", addr)
	return s.Engine.Run(addr)
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Minimal logging for now; we’ll upgrade later.
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		log.Printf("%s %s -> %d", method, path, status)
	}
}
