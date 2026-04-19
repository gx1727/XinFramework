package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/pkg/config"
)

type XinServer struct {
	Engine *gin.Engine
	server *http.Server
}

func New(cfg *config.Config) *XinServer {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	return &XinServer{Engine: engine}
}

func (s *XinServer) Start(addr string) error {
	return s.StartWithSignal(addr, nil)
}

func (s *XinServer) StartWithSignal(addr string, shutdownCh chan<- os.Signal) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.Engine,
	}

	listener, err := newListener(addr)
	if err != nil {
		return fmt.Errorf("create listener failed: %w", err)
	}

	go s.handleSignal(shutdownCh)

	if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server serve failed: %w", err)
	}

	return nil
}

func (s *XinServer) handleSignal(shutdownCh chan<- os.Signal) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	fmt.Printf("\nReceived signal: %v\n", sig)

	if shutdownCh != nil {
		shutdownCh <- sig
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Printf("Graceful shutdown error: %v\n", err)
	}
}

func (s *XinServer) Shutdown(timeout time.Duration) error {
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
