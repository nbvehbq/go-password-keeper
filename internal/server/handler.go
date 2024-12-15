package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		JSONError(res, "hash password", http.StatusInternalServerError)
		return
	}

	userID, err := s.storage.CreateUser(ctx, dto.Login, string(hash))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserExists):
			JSONError(res, err.Error(), http.StatusConflict)
		default:
			JSONError(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	sid, err := s.session.Set(req.Context(), userID)
	if err != nil {
		JSONError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	setCookie(res, sid)
	res.Header().Set("Authorization", sid)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	value := struct {
		SID string `json:"sid"`
	}{SID: sid}

	if err := json.NewEncoder(res).Encode(value); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) loginHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var dto model.RegisterDTO
	if err := json.NewDecoder(req.Body).Decode(&dto); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.storage.GetUserByLogin(ctx, dto.Login)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			JSONError(res, err.Error(), http.StatusUnauthorized)
		default:
			JSONError(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		JSONError(res, err.Error(), http.StatusUnauthorized)
		return
	}

	sid, err := s.session.Set(req.Context(), user.ID)
	if err != nil {
		JSONError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	setCookie(res, sid)
	res.Header().Set("Authorization", sid)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	value := struct {
		SID string `json:"sid"`
	}{SID: sid}

	if err := json.NewEncoder(res).Encode(value); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) createSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var dto model.Secret
	if err := json.NewDecoder(req.Body).Decode(&dto); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.storage.CreateSecret(ctx, &model.Secret{
		UserID:  UID(ctx),
		Type:    dto.Type,
		Payload: dto.Payload,
		Meta:    dto.Meta,
	})
	if err != nil {
		JSONError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	value := struct {
		ID int64 `json:"id"`
	}{ID: id}

	if err := json.NewEncoder(res).Encode(value); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) listSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	type_ := req.URL.Query().Get("type")
	paramType, ok := model.ValidateParam(type_)
	if !ok && type_ != "" {
		JSONError(res, "invalid type", http.StatusBadRequest)
		return
	}

	list, err := s.storage.ListSecrets(ctx, UID(ctx), uint8(paramType))
	if err != nil {
		JSONError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	value := struct {
		Secrets []model.Secret `json:"secrets"`
	}{Secrets: list}

	if err := json.NewEncoder(res).Encode(value); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) getSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	idParam := chi.URLParam(req, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}

	secret, err := s.storage.GetSecret(ctx, int64(id))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrSecretNotFound):
			JSONError(res, err.Error(), http.StatusNotFound)
		default:
			JSONError(res, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(res).Encode(secret); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) updateSecretHandler(res http.ResponseWriter, req *http.Request) {
	var dto model.Secret

	idParam := chi.URLParam(req, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&dto); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := s.storage.UpdateSecret(req.Context(), int64(id), &dto)
	if err != nil {
		JSONError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	value := struct {
		ID int64 `json:"id"`
	}{ID: ret}

	if err := json.NewEncoder(res).Encode(value); err != nil {
		JSONError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) deleteSecretHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	idParam := chi.URLParam(req, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.storage.DeleteSecret(ctx, int64(id))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusNoContent)
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

func UID(ctx context.Context) int64 {
	uid := ctx.Value(uidKey).(int64)
	return uid
}

// JSONError sends an error message in JSON format
func JSONError(w http.ResponseWriter, msg string, code int) {
	res := struct {
		Err string `json:"error"`
	}{Err: msg}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(res)
}
