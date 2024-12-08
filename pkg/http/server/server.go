package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	addr   string
	server *http.Server
}

type Config struct {
	Addr    string
	Handler http.Handler
}

func New(cfg Config) *Server {
	return &Server{
		addr: cfg.Addr,
		server: &http.Server{
			Addr:           cfg.Addr,
			Handler:        cfg.Handler,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    15 * time.Second,
			MaxHeaderBytes: 65536,
			ErrorLog:       nil,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.server.Shutdown(context.Background())
}
