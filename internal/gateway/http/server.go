package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
)

// Server представляет HTTP сервер координатора.
type Server struct {
	gateway gateway.Gateway
	router  *chi.Mux
	server  *http.Server
}

// Config конфигурация HTTP сервера.
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewServer создаёт новый HTTP сервер.
func NewServer(gateway gateway.Gateway, cfg Config) *Server {
	s := &Server{
		gateway: gateway,
	}

	s.router = s.setupRouter()

	s.server = &http.Server{
		Addr:         cfg.Addr,
		Handler:      s.router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return s
}

func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.healthCheck)

		r.Post("/message", s.consumeMessage)

		r.Get("/config", s.getGatewayConfig)
	})

	return r
}

// Start запускает HTTP сервер.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully останавливает сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Router возвращает chi router (для тестов).
func (s *Server) Router() *chi.Mux {
	return s.router
}
