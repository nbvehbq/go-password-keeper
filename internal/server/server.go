package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nbvehbq/go-password-keeper/internal/logger"
	"github.com/nbvehbq/go-password-keeper/internal/model"
)

type Repository interface {
	CreateUser(ctx context.Context, login, pass string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)

	CreateSecret(ctx context.Context, data *model.Secret) (int64, error)
	ListSecrets(ctx context.Context, userID int64, param uint8) ([]model.Secret, error)
	GetSecret(ctx context.Context, id int64) (*model.Secret, error)
	UpdateSecret(ctx context.Context, id int64, data *model.Secret) (int64, error)
	DeleteSecret(ctx context.Context, id int64) error
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

	// Private routes
	r.Group(func(r chi.Router) {
		r.Use(Authenticator(s.session))

		r.Post(`/api/secret`, s.createSecretHandler)
		r.Get(`/api/secret`, s.listSecretHandler)
		r.Get(`/api/secret/{id}`, s.getSecretHandler)
		r.Put(`/api/secret/{id}`, s.updateSecretHandler)
		r.Delete(`/api/secret/{id}`, s.deleteSecretHandler)
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
