package base

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/mad-app/base/blog"
)

// Server provides a server.
type Server struct {
	g       *gin.Engine
	version string
}

func Test() *Server {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

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

func Default() *Server {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	log := blog.Log()

	g.Use(ginzap.RecoveryWithZap(log, true))
	g.Use(ginzap.Ginzap(log, time.RFC3339, true))
	// g.Use(limit.Limit(conf.MaxConcurrentReq))
	// g.Use(limits.RequestSizeLimiter(conf.MaxPostSize))
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

func (r *Server) Serve(port int) error {
	r.g.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "v0.1.0",
		})
	})

	blog.Infof(nil, "%+v", m)

	return r.g.Run(fmt.Sprintf(":%d", port))
}

func (r *Server) Group(path string, gr GroupRouter) *RouteGroup {
	ret := &RouteGroup{
		rg: r.g.Group(path),
		gr: gr,
	}
	m[ret.rg.BasePath()] = gr

	gr.Route(ret)

	return ret
}
