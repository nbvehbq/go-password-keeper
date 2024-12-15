package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nbvehbq/go-password-keeper/internal/logger"
	"github.com/nbvehbq/go-password-keeper/internal/model"
)

type Repository interface {
	CreateUser(ctx context.Context, login, pass string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type SessionStorage interface {
	Set(context.Context, int64) (string, error)
	Get(context.Context, string) (int64, bool)
}

// Server is a keeper server
type Server struct {
	srv     *http.Server
	storage Repository
	session SessionStorage
}

// NewServer creates a new server
func NewServer(storage Repository, session SessionStorage, cfg *Config) (*Server, error) {
	r := chi.NewRouter()

	s := &Server{
		srv:     &http.Server{Addr: cfg.Address, Handler: r},
		session: session,
		storage: storage,
	}

	r.Use(logger.Middleware)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post(`/api/user/register`, s.registerHandler)
		r.Post(`/api/user/login`, s.loginHandler)
	})

	r.Mount("/debug", middleware.Profiler())

	return s, nil
}

// Run runs the server
func (s *Server) Run() error {
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
