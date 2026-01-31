package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
)

// Server представляет HTTP сервер координатора.
type Server struct {
	coordinator coordinator.Coordinator
	router      *chi.Mux
	server      *http.Server
}

// Config конфигурация HTTP сервера.
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewServer создаёт новый HTTP сервер.
func NewServer(coordinator coordinator.Coordinator, cfg Config) *Server {
	s := &Server{
		coordinator: coordinator,
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
		// Health check
		r.Get("/health", s.healthCheck)

		// Nodes
		r.Route("/nodes", func(r chi.Router) {
			r.Post("/", s.createNode)
			r.Get("/", s.listNodes)
			r.Get("/{nodeID}", s.getNode)
			r.Put("/{nodeID}/heartbeat", s.updateNodeHeartbeat)
			r.Delete("/{nodeID}", s.deleteNode)
		})

		// Gateways
		r.Route("/gateways", func(r chi.Router) {
			r.Post("/", s.registerGateway)
			r.Get("/", s.listGateways)
			r.Get("/{gatewayID}", s.getGateway)
			r.Put("/{gatewayID}/heartbeat", s.updateGatewayHeartbeat)
			r.Delete("/{gatewayID}", s.unregisterGateway)
		})

		// Storages
		r.Route("/storages", func(r chi.Router) {
			r.Post("/", s.registerStorage)
			r.Get("/", s.listStorages)
			r.Get("/{storageID}", s.getStorage)
			r.Put("/{storageID}/heartbeat", s.updateStorageHeartbeat)
			r.Delete("/{storageID}", s.unregisterStorage)
		})

		// Pushers
		r.Route("/pushers", func(r chi.Router) {
			r.Post("/", s.registerPusher)
			r.Get("/", s.listPushers)
			r.Get("/{pusherID}", s.getPusher)
			r.Put("/{pusherID}/heartbeat", s.updatePusherHeartbeat)
			r.Delete("/{pusherID}", s.unregisterPusher)
		})

		// Routing Rules
		r.Route("/routing-rules", func(r chi.Router) {
			r.Post("/", s.createRoutingRule)
			r.Get("/", s.listRoutingRules)
			r.Get("/{ruleID}", s.getRoutingRule)
			r.Put("/{ruleID}", s.updateRoutingRule)
			r.Delete("/{ruleID}", s.deleteRoutingRule)
		})

		// Config (read-only)
		r.Get("/coordinator-config", s.getCoordinatorConfig)
		r.Get("/cluster-config", s.getClusterConfig)
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
