package base

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server provides a server.
type Server struct {
	g       *gin.Engine
	version string
}

func Default() *Server {
	g := gin.Default()

	g.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowWebSockets:  true,
		AllowAllOrigins:  true,
		MaxAge:           12 * time.Hour,
	}))

	return &Server{
		g: g,
	}
}

func (r *Server) Serve() error {
	r.g.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "v0.1.0",
		})
	})

	return r.g.Run()
}

func (r *Server) Group(path string, gr GroupRouter) *RouteGroup {
	ret := &RouteGroup{
		rg: r.g.Group(path),
		gr: gr,
	}

	gr.Route(ret)

	return ret
}
