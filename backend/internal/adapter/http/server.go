package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	server *http.Server
}

func NewServer(addr string, engine *gin.Engine) *Server {
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: engine,
		},
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
