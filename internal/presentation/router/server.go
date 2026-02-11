package router

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Server struct {
	e    *echo.Echo
	port int64
}

func NewServer(e *echo.Echo, port int64) *Server {
	return &Server{
		e,
		port,
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.e.Start(fmt.Sprintf(":%d", s.port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error starting server. %v. Shutting down now", err)
		}
	}()
}

// Shutdown shuts down the server.
func (s *Server) Shutdown(ctx context.Context) {
	if err := s.e.Shutdown(ctx); err != nil {
		log.Fatalf("error shutting down server: %v", err)
	}
}
