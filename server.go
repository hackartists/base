package base

import (
	"log"
	"net/http"
	"time"
)

// Server provides a server.
type Server struct {
	s *http.Server
}

func Default() *Server {
	return &Server{
		s: &http.Server{
			Addr:           ":8080",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

func (r *Server) Serve() error {
	log.Fatal(r)
	return r.s.ListenAndServe()
}
