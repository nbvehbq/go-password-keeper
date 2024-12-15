package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nbvehbq/go-password-keeper/internal/logger"
	"github.com/nbvehbq/go-password-keeper/internal/model"
	"github.com/nbvehbq/go-password-keeper/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerHandler(res http.ResponseWriter, req *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Log.Error("error", zap.Error(err))
		}
	}()

	ctx := req.Context()

	var dto model.RegisterDTO
	if err = json.NewDecoder(req.Body).Decode(&dto); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(res, "hash password", http.StatusInternalServerError)
		return
	}

	userID, err := s.storage.CreateUser(ctx, dto.Login, string(hash))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserExists):
			http.Error(res, err.Error(), http.StatusConflict)
		default:
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	sid, err := s.session.Set(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	setCookie(res, sid)
	res.Header().Set("Authorization", sid)

	res.WriteHeader(http.StatusOK)
}

func (s *Server) loginHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var dto model.RegisterDTO
	if err := json.NewDecoder(req.Body).Decode(&dto); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.storage.GetUserByLogin(ctx, dto.Login)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			http.Error(res, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		http.Error(res, "", http.StatusUnauthorized)
		return
	}

	sid, err := s.session.Set(req.Context(), user.ID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	setCookie(res, sid)
	res.Header().Set("Authorization", sid)

	res.WriteHeader(http.StatusOK)
}

func setCookie(w http.ResponseWriter, payload string) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    payload,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}
