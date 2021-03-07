package base

import (
	"github.com/gin-gonic/gin"
)

// Server provides a server.
type Server struct {
	g       *gin.Engine
	version string
}

func Default() *Server {
	g := gin.Default()

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
