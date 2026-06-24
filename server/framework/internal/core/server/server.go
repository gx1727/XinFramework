package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/config"
)

type Server struct {
	Engine *gin.Engine
	server *http.Server
}

func New(cfg *config.Config) *Server {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	engine.Static(cfg.Storage.LocalBaseURL, cfg.Storage.LocalDir)

	return &Server{Engine: engine}
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}

	listener, err := newListener(addr)
	if err != nil {
		return fmt.Errorf("create listener failed: %w", err)
	}

	if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server serve failed: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(timeout time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func newListener(addr string) (*netListen, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &netListen{TCPListener: ln.(*net.TCPListener)}, nil
}

type netListen struct {
	*net.TCPListener
}

func (ln netListen) Accept() (net.Conn, error) {
	return ln.TCPListener.Accept()
}

func (ln netListen) Close() error {
	return ln.TCPListener.Close()
}
