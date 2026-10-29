package server

import (
	"context"
	"github.com/codewec/uptimelit/backend/internal/auth"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/codewec/uptimelit/backend/internal/db"
	"github.com/codewec/uptimelit/backend/internal/monitor"
)

type oidcAuthenticator interface {
	AuthorizationURL(state, nonce, verifier string) string
	Authenticate(context.Context, string, string, string) (*auth.OIDCIdentity, error)
}

type Server struct {
	router     *chi.Mux
	db         *db.DB
	manager    *monitor.Manager
	auth       *auth.Service
	authConfig auth.Config
	oidc       oidcAuthenticator
	limiter    limiter
	frontend   fs.FS
}

func New(database *db.DB, manager *monitor.Manager, cfg auth.Config, frontend fs.FS) (*Server, error) {
	svc, err := auth.New(database, cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	oidc, err := auth.NewOIDCAuthenticator(ctx, &cfg)
	if err != nil {
		return nil, err
	}
	if cfg.OIDCInsecureSkipTLSVerify {
		slog.Warn("OIDC TLS certificate verification is disabled")
	}
	s := &Server{router: chi.NewRouter(), db: database, manager: manager, auth: svc, authConfig: cfg, frontend: frontend}
	if oidc != nil {
		s.oidc = oidc
	}
	s.router.Use(middleware.Logger, middleware.Recoverer)
	s.routes()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) routes() {
	s.router.Get("/*", s.serveFrontend)

	s.router.Route("/api", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"), s.authentication)
		s.authRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(requireAuth, s.resourceOwner)

			// System Status
			r.Get("/status", s.getStatus)

			// Endpoints
			r.Get("/endpoints", s.listEndpoints)
			r.Post("/endpoints", s.createEndpoint)
			r.Get("/endpoints/{id}", s.getEndpoint)
			r.Put("/endpoints/{id}", s.updateEndpoint)
			r.Delete("/endpoints/{id}", s.deleteEndpoint)
			r.Post("/endpoints/{id}/pause", s.pauseEndpoint)
			r.Get("/endpoints/{id}/checks", s.listChecks)
			r.Get("/endpoints/{id}/history", s.history)

			// Notifiers
			r.Get("/notifiers", s.listNotifiers)
			r.Post("/notifiers", s.createNotifier)
			r.Get("/notifiers/{id}", s.getNotifier)
			r.Put("/notifiers/{id}", s.updateNotifier)
			r.Delete("/notifiers/{id}", s.deleteNotifier)
			r.Post("/notifiers/{id}/test", s.testNotifier)
		})
	})
}
